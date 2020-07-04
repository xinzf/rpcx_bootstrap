package bootstrap

//
//import (
//    "net"
//    "github.com/smallnest/rpcx/serverplugin"
//    "io/ioutil"
//    "fmt"
//)
//
//type FileUploadRequest struct {
//    Body []byte
//    FileName string
//    FileSize int64
//    Meta map[string]string
//}
//
//type UploaderHandler func(req *FileUploadRequest,rsp interface{}) error
//
//
//
//func RegisterUploader(name string, handler UploaderHandler) error {
//    l, _ := net.Listen("tcp", ":0")
//    port := l.Addr().(*net.TCPAddr).Port
//    l.Close()
//
//    addr:=fmt.Sprintf("%s:%d",Config.Server.Host,port)
//    Logger.Info("upload handler listen on:",addr)
//    p := serverplugin.NewFileTransfer(addr, func(conn net.Conn, args *serverplugin.FileTransferArgs) {
//        data, err := ioutil.ReadAll(conn)
//        if err != nil {
//            Logger.Error("upload file err:",err)
//            return
//        }
//
//        req := &FileUploadRequest{
//            Body:data,
//            FileName:args.FileName,
//            FileSize:args.FileSize,
//            Meta:args.Meta,
//        }
//        if err := handler(req); err != nil {
//            fmt.Printf("error read: %v\n", err)
//            return
//        }
//
//        fmt.Println("Upload success")
//    }, nil, 1000)
//    serverplugin.RegisterFileTransfer(server, p)
//    return nil
//}
