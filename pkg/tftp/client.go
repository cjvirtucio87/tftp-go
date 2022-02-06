package tftp

type TftpClient interface {
  // send a read request to a server at the given address and port
  Send(req ReadRequest, addr string) error
}
