package tools

import "errors"

var errContainerCreate = errors.New("failed to create container")
var errContainerStart = errors.New("failed to start container")
var errContainerAttach = errors.New("failed to attach to container")
var errContainerWait = errors.New("failed to wait on container")
var errContainerLogRead = errors.New("failed to read container logs")
var errTerminalConvert = errors.New("failed to convert terminal to raw")
var errContainerInputSend = errors.New("failed to send input to container")
var errRuntimeInfoFetch = errors.New("failed to fetch runtime info")
