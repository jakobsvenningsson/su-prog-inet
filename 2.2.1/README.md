# 2.2.1 Datagram sockets med unicast

This assignment implements a distributed whiteboard. The UI (frontend) of the application is implemented as a web application and the backend server in golang.

Start the application: `go run main.go {port} {remoteport}` where "port" is the port which your application will use for incoming UDP traffic and "remotePort" is a comma-seperated list of remote clients. The UI can be reached on `localhost:{port}`.

The following example script starts three clients:
`go run main.go 5000 5001,5002 &
go run main.go 5001 5000,5002 &
go run main.go 5002 5000,5001`


