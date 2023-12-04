package main

//func TestCreateNode(t *testing.T) {
//	listenAddr := "/ip4/0.0.0.0/tcp/0"
//	node1, err := CreateNode(listenAddr)
//	if err != nil {
//		fmt.Println("Error creating node:", err)
//		return
//	}
//
//	node2, err := CreateNode(listenAddr)
//	if err != nil {
//		fmt.Println("Error creating node:", err)
//		return
//	}
//
//	// 等待 SIGINT 或 SIGTERM 信号
//	ch := make(chan os.Signal, 1)
//	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
//	<-ch
//	fmt.Println("Received signal, shutting down...")
//
//	// 关闭节点
//	if err := (*node1).Close(); err != nil {
//		fmt.Println("Error closing node:", err)
//	}
//	if err := (*node2).Close(); err != nil {
//		fmt.Println("Error closing node:", err)
//	}
//}
