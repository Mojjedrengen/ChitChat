:::: Chit Chat ::::


In this assignment you will design and implement Chit Chat, a

distributed chat service where participants can join, exchange

messages, and leave the conversation at any time. Chit Chat is a

lively playground for exploring the essence of distributed systems:

communication, coordination, and the ordering of events in a world

without a single shared clock.



: System Specification : 


The system must satisfy the following specifications:


S1. Chit Chat is a distributed service that enables clients to

exchange chat messages using gRPC for all communication. Students must

design the gRPC API, including all service methods and message types.


S2. The Chit Chat system must follow a distributed topology consisting

of one service process and multiple client processes. Each client runs

as an independent process that communicates with the service via

gRPC. A minimal configuration must include one service instance and at

least three concurrently active clients.


S3. Each participant can publish a valid chat message at any time. A

valid message is a UTF-8 encoded string with a maximum length of 128

characters. Publishing is performed through a gRPC call to the Chit

Chat service.


S4. The Chit Chat service must broadcast each published message to all

currently active participants. Each broadcast must include the message

content and a logical timestamp.


S5. Participants may join the system at any time.  When a new

participant X joins, the service must broadcast a message of the form:

Participant X joined Chit Chat at logical time L. This message must

be delivered to all participants, including the newly joined one.


S6. Participants may leave the system at any time.  When a participant

X leaves, the service must broadcast a message of the form:

Participant X left Chit Chat at logical time L. This message must be

delivered to all remaining participants.


S7. When a participant receives any broadcast message, it must: i)

display the message content and its logical timestamp on the client

interface; and ii) log the message content and its logical timestamp.




:Technical Requirements:


- The system must be implemented in Go


- The gRPC framework must be used for client–server communication,

  using Protocol Buffers (you must provide a .proto file) for message

  definitions.


- The Go log standard library must be used for structured logging of

  events (e.g. client join/leave notifications, message delivery, and

  server startup/shutdown).


- Concurrency (local to the server or the clients) must be handled

  using Go routines and channels for synchronised communication

  between components


- Every client and the server must be deployed as separate processes


- Each client connection must be served by a dedicated goroutine

  managed by the server.


- The system must support multiple concurrent client connections

  without blocking message delivery.


- The system must log the following events:

  * Server startup and shutdown

  * Client connection and disconnection events

  * Broadcast of join/leave messages

  * Message delivery events


- Log messages must include:

  * Timestamp

  * Component name (Server/Client)

  * Event type

  * Relevant identifiers (e.g. Client ID).


- The system can be started with at least three (3) nodes (two client

  and a server) and it must be able to handle join of at least one

  client and leave of at least one client




Hand-in Requirements


- you must hand in a report (single pdf file) via LearnIT


- you must provide a link to a Git repo with your source code in the

  report


- in the report, you must


  * discuss, whether you are going to use server-side streaming,

    client-side streaming, or bidirectional streaming?


  * describe your system architecture - do you have a server-client

    architecture, peer-to-peer, or something else?


  * describe what RPC methods are implemented, of what type, and what

    messages types are used for communication

  

  * describe how you have implemented the calculation of the timestamps

  

  * provide a diagram, that traces a sequence of RPC calls together

    with the Lamport timestamps, that corresponds to a chosen sequence

    of interactions: Client X joins, Client X Publishes, ..., Client X

    leaves. 


- you must include system logs that document the requirements are met,

    in both the appendix of your report and your repo


- your repo must include a readme.md file that describes how to run

  your program.


- you repo must structured as follows:
