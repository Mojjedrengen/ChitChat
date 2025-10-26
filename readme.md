> [!IMPORTANT]
> The ui is written for Linux using ANSI Escape Squences and might not work in Windows powershell.
> We cannot garante that the program runs as intented on Windows
> It is therefor highly recommened that the program is run on Linux or WSL.

To use run the server program `server/server.go` and the boot up as many clients as you want `client/client.go`
To write a message as a client just write it.
To Disconnect from the server as a client call SIGTERM, control+c.

When starting the programs you can parse a number of flags.
To use this wiret `{program} -{flag} {value}`
| Flag | Type | Default | Desc |
| --- | --- | --- | ---|
| addr | string | localhost:50051 | The server adress in the format of host:port |
| folder | string | data/ | The folder for where the log is saved |

**Table 1:** Clients flags

| Flag   | Type   | Default   | Desc                                                   |
| ------ | ------ | --------- | ------------------------------------------------------ |
| port   | int    | 50051     | The server port                                        |
| file   | string | data.json | Name of file where the message history is stored       |
| folder | string | data      | the folder the the message history and logs are stored |

**Table 2:** Servers flags

Missing features:

# Bugs:
The client does not read Windows termination signals, can't figure out why and how to fix it, so the the client can't use the disconnect funtion of Windows.
This means that the server does not get the message that the client is disconnected and so it does not brodcast it.
To mitigate this run the program on Linux

used [GRPC.Go.Basic_turtorial](https://grpc.io/docs/languages/go/basics/) as a refrence to create the project
