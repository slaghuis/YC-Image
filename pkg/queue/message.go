package queue

type ImageJob struct {
    ID       string `json:"id"`
    Category string `json:"category"`
    OrigPath string `json:"orig_path"`
}
