package sandbox

// const GCCImageTags = "gcc:latest"

// func CheckGCCImage() {
// 	if !existGCCImage() {
// 		PullGCCImage()
// 	}
// }

// func existGCCImage() bool {
// 	ctx := context.Background()
// 	images, err := cli.ImageList(ctx, types.ImageListOptions{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, image := range images {
// 		if image.RepoTags[0] == GCCImageTags {
// 			return true
// 		}
// 	}
// 	return false
// }

// type pullStatus struct {
// 	Status   string `json:"status"`
// 	Progress string `json:"progress"`
// }

// func (s *pullStatus) Write(p []byte) (n int, err error) {
// 	json.Unmarshal(p, s)
// 	fmt.Printf(s.Status + ": " + s.Progress + "\r")
// 	return len(p), nil
// }

// func PullGCCImage() {
// 	log.Printf("pulling gcc image...\n")
// 	ctx := context.Background()
// 	reader, err := cli.ImagePull(ctx, GCCImageTags, types.ImagePullOptions{})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer reader.Close()
// 	var status pullStatus

// 	io.Copy(&status, reader)
// }
