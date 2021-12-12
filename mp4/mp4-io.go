package mp4

type Reader interface {
}

type Writer interface {
    Write(p []byte) (n int, err error)
    Seek(offset int64, whence int) (int64, error)
    Tell() (offset int64)
}
