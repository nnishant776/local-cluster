//go:build nobuild

package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/nnishant776/local-cluster/config"
	"github.com/spf13/cobra"
)

const (
	k3sGitAccount = "k3s-io"
	k3sGitRepo    = "k3s"
)

func init() {

}

func cpuArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "x86_64":
		arch = "amd64"
	default:
		if strings.Contains(arch, "arm") {
			arch = "arm"
		}
	}
	return arch
}

func k3sBinaryName() string {
	assetName := "k3s"
	arch := runtime.GOARCH
	switch arch {
	case "x86_64", "amd64":
	default:
		if strings.Contains(arch, "arm") {
			arch = "armhf"
		}

		assetName += "-" + arch
	}

	return assetName
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "prebuild",
		Short:            "prebuild is a program to prep the build for lcctl",
		Long:             "prebuild is a program to prep the build for lcctl",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
	}

	rootCmd.PersistentFlags().Bool("verbose", false, "Print verbose errors")

	return rootCmd
}

func downloadAsset(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	relAsset *github.ReleaseAsset,
) (io.ReadCloser, error) {
	content, redirect, err := client.Repositories.DownloadReleaseAsset(
		ctx, k3sGitAccount, k3sGitRepo, relAsset.GetID(), http.DefaultClient,
	)
	if err != nil {
		slog.Info("Failed to download release asset", "error", err)
		return nil, err
	}

	if content == nil {
		slog.Info("Fetching asset content from the redirect URL", "url", redirect)
		req, err := http.NewRequest(http.MethodGet, redirect, nil)
		if err != nil {
			slog.Error("Failed to query redirect URL", "error", err)
			return nil, err
		}

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Failed to perform HTTP request", "error", err)
			return nil, err
		}

		content = response.Body
	}

	return content, nil
}

func findReleaseAsset(name string, assets []*github.ReleaseAsset) *github.ReleaseAsset {
	for _, asset := range assets {
		if *asset.Name == name {
			return asset
		}
	}

	return nil
}

func newDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "download",
		Short:            "download k3s binary from GitHub release",
		Long:             "download k3s binary from GitHub release",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get k8s version
			tag := cmd.Flag("tag").Value.String() + "+k3s1"
			slog.Info("Fetching release info", "tag", tag)

			// Get the release information
			ghClient := github.NewClient(http.DefaultClient)
			relInfo, _, err := ghClient.Repositories.GetReleaseByTag(
				cmd.Context(), k3sGitAccount, k3sGitRepo, tag,
			)
			if err != nil {
				slog.Error("Failed to fetch release info", "tag", tag, "error", err)
				return err
			}

			// Find the correct asset name
			assetName := k3sBinaryName()
			relAsset := findReleaseAsset(assetName, relInfo.Assets)
			if relAsset != nil {
				slog.Info("Found matching release asset", "name", assetName, "id", relAsset.GetID())
			} else {
				slog.Error("Matching asset not found")
				return fmt.Errorf("asset '%s' not found", assetName)
			}

			// Obtain a download handle for the asset
			k3sBinContent, err := downloadAsset(cmd.Context(), ghClient, k3sGitAccount, k3sGitRepo, relAsset)
			if err != nil {
				return err
			}
			defer k3sBinContent.Close()

			// Create the file to download the asset
			path := cmd.Flag("path").Value.String()
			file, err := os.Create(filepath.Join(path, "k3s"))
			if err != nil {
				slog.Error("Failed to open file", "name", path, "error", err)
				return err
			}
			defer file.Close()

			// Check if checksum verification is enabled and update write handle accordingly
			verifyChecksum := cmd.Flag("verify").Value.String() == "true"
			hasher := sha256.New()
			dst := (io.Writer)(file)
			if verifyChecksum {
				dst = io.MultiWriter(file, hasher)
			}

			// Wrap the download handle for progress report
			if cmd.Flag("report-progress").Value.String() == "true" {
				pr := newProgressReportReader(
					k3sBinContent, relAsset.GetSize(), relAsset.GetSize()/100, 500*time.Millisecond,
				)
				pr.logFunc = func(read int, total int) {
					if total > 0 {
						fmt.Printf("\rTransferred [%s]: %d out of %d bytes", *relAsset.Name, read, total)
					} else {
						fmt.Printf("\rTransferred [%s]: %d bytes", *relAsset.Name, read, total)
					}
				}
				k3sBinContent = io.NopCloser(pr)
			}

			// Start downloading the file
			_, err = io.Copy(dst, k3sBinContent)
			if err != nil {
				slog.Error("Failed to write file", "name", path, "error", err)
				return err
			}

			fmt.Println()

			if verifyChecksum {
				// Get the checksum file from the release assets
				chksumFilename := "sha256sum-" + cpuArch() + ".txt"
				chksumAsset := findReleaseAsset(chksumFilename, relInfo.Assets)
				if chksumAsset == nil {
					slog.Error("Checksum file not found", "arch", cpuArch())
					return fmt.Errorf("asset '%s' not found", chksumFilename)
				}

				// Obtain the download handle for the checksum file
				chksumContent, err := downloadAsset(cmd.Context(), ghClient, k3sGitAccount, k3sGitRepo, chksumAsset)
				if err != nil {
					return err
				}
				defer chksumContent.Close()

				// Read checksum file to get original sha256 checksum for the downloaded binary
				scanner := bufio.NewScanner(chksumContent)
				chksumOrig := ""
				for scanner.Scan() {
					line := scanner.Text()
					sum, remainder, found := strings.Cut(line, " ")
					if !found {
						break
					}
					remainder = strings.TrimSpace(remainder)
					if remainder == *relAsset.Name {
						chksumOrig = sum
						break
					}
				}

				// Compare downloaded file checksum with the original checksum
				chksumDown := hex.EncodeToString(hasher.Sum(nil))
				if chksumOrig != chksumDown {
					return fmt.Errorf("checksum mismatch: expected=%s, actual=%s", chksumOrig, chksumDown)
				}

				slog.Info("Checksum verified successfully. K3S dowload completed")
			}

			return nil
		},
	}

	cmd.Flags().Bool("verify", true, "Verify the binary checksum")
	cmd.Flags().String("tag", config.GetK8SVersion(), "Specify the release to download")
	cmd.Flags().String("path", ".", "Specify the path where the release will be downloaded")
	cmd.Flags().Bool("report-progress", true, "Specify whether to report download progress")

	return cmd
}

func newProgressReportReader(
	src io.Reader, totalBytes, minReportBytes int, reportInterval time.Duration,
) *progressReportTransformer {
	return &progressReportTransformer{
		src:                src,
		minReportBytes:     minReportBytes,
		totalExpectedBytes: totalBytes,
		lastReport:         time.Now(),
		reportInterval:     reportInterval,
	}
}

type progressReportTransformer struct {
	src                io.Reader
	minReportBytes     int
	totalReadBytes     int
	totalExpectedBytes int
	lastReport         time.Time
	lastUpdateBytes    int
	reportInterval     time.Duration
	logFunc            func(read, total int)
}

func (self *progressReportTransformer) Read(b []byte) (int, error) {
	n, err := self.src.Read(b)
	self.totalReadBytes += n
	t := time.Now()

	if self.minReportBytes > 0 {
		if self.totalReadBytes-self.lastUpdateBytes >= self.minReportBytes || self.totalReadBytes == self.totalExpectedBytes {
			self.logFunc(self.totalReadBytes, self.totalExpectedBytes)
			self.lastUpdateBytes = self.totalReadBytes
			self.lastReport = t
		}
	}

	if self.reportInterval > 0 && time.Now().Sub(self.lastReport) >= self.reportInterval {
		self.logFunc(self.totalReadBytes, self.totalExpectedBytes)
		self.lastUpdateBytes = self.totalReadBytes
		self.lastReport = t
	}

	return n, err
}

func main() {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newDownloadCmd())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
