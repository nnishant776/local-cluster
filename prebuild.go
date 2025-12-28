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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/nnishant776/local-cluster/config"
	"github.com/spf13/cobra"
)

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
	downloadList := []*cobra.Command{}
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			component := cmd.Flag("component").Value.String()
			switch component {
			case "k3s":
				downloadList = append(downloadList, newK3SDownloadCmd())
			case "k9s":
				downloadList = append(downloadList, newK9SDownloadCmd())
			case "kubectl":
				downloadList = append(downloadList, newKubectlDownloadCmd())
			case "helm":
				downloadList = append(downloadList, newHelmDownloadCmd())
			case "helmfile":
				downloadList = append(downloadList, newHelmfileDownloadCmd())
			}

			for _, c := range downloadList {
				cmd.AddCommand(c)
			}

			cmd.ParseFlags(args)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.RunE, cmd.PreRunE = nil, nil
			return cmd.ExecuteContext(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().String("component", "all", "Specify the component to download")
	rootCmd.PersistentFlags().Bool("verbose", false, "Print verbose errors")
	rootCmd.PersistentFlags().Bool("report-progress", true, "Specify whether to report download progress")
	rootCmd.PersistentFlags().String("path", ".", "Specify the path where the release will be downloaded")
	rootCmd.PersistentFlags().String("tag", "latest", "Specify the application version")

	return rootCmd
}

func downloadGithubReleaseAsset(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	relAsset *github.ReleaseAsset,
) (io.ReadCloser, error) {
	content, redirect, err := client.Repositories.DownloadReleaseAsset(
		ctx, owner, repo, relAsset.GetID(), http.DefaultClient,
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

func downloadFileFromURL(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Info("Failed to construct URL", "error", err)
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Info("Failed to perform HTTP call", "error", err)
		return nil, err
	}

	return res, nil
}

func findReleaseAsset(name string, assets []*github.ReleaseAsset) *github.ReleaseAsset {
	for _, asset := range assets {
		if *asset.Name == name {
			return asset
		}
	}

	return nil
}

func newK3SDownloadCmd() *cobra.Command {
	const (
		k3sGitAccount = "k3s-io"
		k3sGitRepo    = "k3s"
	)

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
			tag := config.GetK8SVersion() + "+k3s1"
			slog.Info("Fetching release info for k3s", "tag", tag)

			// Get the release information
			ghClient := github.NewClient(http.DefaultClient)
			relInfo, _, err := ghClient.Repositories.GetReleaseByTag(
				cmd.Context(), k3sGitAccount, k3sGitRepo, tag,
			)
			if err != nil {
				slog.Error("Failed to fetch release info", "tag", tag, "error", err)
				return err
			}

			// ghClient = ghClient.WithAuthToken(cmp.Or(os.Getenv("GH_TOKEN"), os.Getenv("GITHUB_TOKEN")))

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
			k3sBinContent, err := downloadGithubReleaseAsset(cmd.Context(), ghClient, k3sGitAccount, k3sGitRepo, relAsset)
			if err != nil {
				return err
			}
			defer k3sBinContent.Close()

			// Create the file to download the asset
			path := cmd.Flag("path").Value.String()
			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return err
			}
			file, err := os.OpenFile(filepath.Join(path, "k3s"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				slog.Error("Failed to open file", "name", path, "error", err)
				return err
			}
			defer file.Close()

			// Check if checksum verification is enabled and update write handle accordingly
			hasher := sha256.New()
			dst := io.MultiWriter(file, hasher)

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

			// Get the checksum file from the release assets
			chksumFilename := "sha256sum-" + cpuArch() + ".txt"
			chksumAsset := findReleaseAsset(chksumFilename, relInfo.Assets)
			if chksumAsset == nil {
				slog.Error("Checksum file not found", "arch", cpuArch())
				return fmt.Errorf("asset '%s' not found", chksumFilename)
			}

			// Obtain the download handle for the checksum file
			chksumContent, err := downloadGithubReleaseAsset(cmd.Context(), ghClient, k3sGitAccount, k3sGitRepo, chksumAsset)
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

			return nil
		},
	}

	return cmd
}

func newKubectlDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "download",
		Short:            "download kubectl binary from GitHub release",
		Long:             "download kubectl binary from GitHub release",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get k8s version
			tag := config.GetK8SVersion()
			slog.Info("Fetching release info for kubectl", "tag", tag)

			releaseURL := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/%s/kubectl", tag, runtime.GOARCH)

			// Obtain a download handle for the asset
			kubectlBinResponse, err := downloadFileFromURL(cmd.Context(), releaseURL)
			if err != nil {
				return err
			}

			kubectlBinContent := kubectlBinResponse.Body
			defer kubectlBinContent.Close()

			// Create the file to download the asset
			path := cmd.Flag("path").Value.String()
			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return err
			}
			file, err := os.OpenFile(filepath.Join(path, "kubectl"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				slog.Error("Failed to open file", "name", path, "error", err)
				return err
			}
			defer file.Close()

			// Check if checksum verification is enabled and update write handle accordingly
			hasher := sha256.New()
			dst := io.MultiWriter(file, hasher)

			// Wrap the download handle for progress report
			if cmd.Flag("report-progress").Value.String() == "true" {
				pr := newProgressReportReader(
					kubectlBinContent, int(kubectlBinResponse.ContentLength), int(kubectlBinResponse.ContentLength)/100, 500*time.Millisecond,
				)
				pr.logFunc = func(read int, total int) {
					if total > 0 {
						fmt.Printf("\rTransferred [%s]: %d out of %d bytes", "kubectl", read, total)
					} else {
						fmt.Printf("\rTransferred [%s]: %d bytes", "kubectl", read, total)
					}
				}
				kubectlBinContent = io.NopCloser(pr)
			}

			// Start downloading the file
			_, err = io.Copy(dst, kubectlBinContent)
			if err != nil {
				slog.Error("Failed to write file", "name", path, "error", err)
				return err
			}

			fmt.Println()

			// Get the checksum file from the release assets
			releaseChecksum := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/%s/kubectl.sha256", tag, runtime.GOARCH)

			// Obtain the download handle for the checksum file
			checksumResponse, err := downloadFileFromURL(cmd.Context(), releaseChecksum)
			if err != nil {
				return err
			}
			chksumContent := checksumResponse.Body
			defer chksumContent.Close()

			// Read checksum file to get original sha256 checksum for the downloaded binary
			chksumBytes, err := io.ReadAll(chksumContent)
			if err != nil {
				slog.Error("Failed to read checksum data", "error", err)
				return err
			}

			// Compare downloaded file checksum with the original checksum
			chksumDown := hex.EncodeToString(hasher.Sum(nil))
			if string(chksumBytes) != chksumDown {
				return fmt.Errorf("checksum mismatch: expected=%s, actual=%s", string(chksumBytes), chksumDown)
			}

			slog.Info("Checksum verified successfully. Kubectl dowload completed")

			return nil
		},
	}

	return cmd
}

func newK9SDownloadCmd() *cobra.Command {
	const (
		k9sGitAccount = "derailed"
		k9sGitRepo    = "k9s"
	)

	cmd := &cobra.Command{
		Use:              "download",
		Short:            "download k9s binary from GitHub release",
		Long:             "download k9s binary from GitHub release",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get k9s version
			tag := cmd.Flag("tag").Value.String()
			path := cmd.Flag("path").Value.String()
			path, err := filepath.Abs(path)
			if err != nil {
				slog.Error("Failed to get absolute path", "input", path)
				return err
			}
			modName := fmt.Sprintf("github.com/%s/%s@%s", k9sGitAccount, k9sGitRepo, tag)
			slog.Info("Fetching release info for k9s", "tag", tag, "url", modName)
			proc := exec.Command("go", "install", "-ldflags", "-s -w", modName)
			proc.Env = append(os.Environ(), "GOBIN="+path)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			err = proc.Run()
			if err != nil {
				slog.Error("Failed to download k9s", "error", err)
				return err
			}

			return nil
		},
	}

	return cmd
}

func newHelmDownloadCmd() *cobra.Command {
	const (
		helmGitAccount = "helm.sh"
		helmGitRepo    = "helm/v3/cmd/helm"
	)

	cmd := &cobra.Command{
		Use:              "download",
		Short:            "download helm binary from GitHub release",
		Long:             "download helm binary from GitHub release",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get k9s version
			tag := cmd.Flag("tag").Value.String()
			path := cmd.Flag("path").Value.String()
			path, err := filepath.Abs(path)
			if err != nil {
				slog.Error("Failed to get absolute path", "input", path)
				return err
			}
			modName := fmt.Sprintf("%s/%s@%s", helmGitAccount, helmGitRepo, tag)
			slog.Info("Fetching release info for helm", "tag", tag, "url", modName)
			proc := exec.Command("go", "install", "-ldflags", "-s -w", modName)
			proc.Env = append(os.Environ(), "GOBIN="+path)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			err = proc.Run()
			if err != nil {
				slog.Error("Failed to download helm", "error", err)
				return err
			}

			return nil
		},
	}

	return cmd
}

func newHelmfileDownloadCmd() *cobra.Command {
	const (
		helmfileGitAccount = "helmfile"
		helmfileGitRepo    = "helmfile"
	)

	cmd := &cobra.Command{
		Use:              "download",
		Short:            "download helmfile binary from GitHub release",
		Long:             "download helmfile binary from GitHub release",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get k9s version
			tag := cmd.Flag("tag").Value.String()
			path := cmd.Flag("path").Value.String()
			path, err := filepath.Abs(path)
			if err != nil {
				slog.Error("Failed to get absolute path", "input", path)
				return err
			}
			modName := fmt.Sprintf("github.com/%s/%s@%s", helmfileGitAccount, helmfileGitRepo, tag)
			slog.Info("Fetching release info for helmfile", "tag", tag, "url", modName)
			proc := exec.Command("go", "install", "-ldflags", "-s -w", modName)
			proc.Env = append(os.Environ(), "GOBIN="+path)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			err = proc.Run()
			if err != nil {
				slog.Error("Failed to download helmfile", "error", err)
				return err
			}

			return nil
		},
	}

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

func (self *progressReportTransformer) readHook(bytesRead int) {
	self.totalReadBytes += bytesRead
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
}

func (self *progressReportTransformer) Read(b []byte) (int, error) {
	n, err := self.src.Read(b)
	self.readHook(n)
	return n, err
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		panic(err)
	}
}
