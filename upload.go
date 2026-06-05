package main

type Uploader interface {
	Upload(localPath, host, remotePath string) error
}
