# tcp mockserver
This server acts like a mock for TCP services.
It can send a preamble after the client connects, and then responds to received messages.
The responses consist of sequences of `[]byte` with configurable delays.

## Todo
make the sequence a list of sequences?
