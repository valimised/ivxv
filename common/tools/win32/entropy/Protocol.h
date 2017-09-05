#pragma once

#include <winsock2.h>

#define DATABUFSIZE 8192

class Protocol {

public:

    Protocol(SOCKET);
    virtual ~Protocol();

    void handle(DWORD bytesTransferred);
    bool recv();

protected:

    OVERLAPPED overlapped;
    SOCKET socket;
    CHAR buffer[DATABUFSIZE];
    WSABUF dataBuf;
    DWORD bytesSEND;
    DWORD bytesTOBESENT;
    DWORD bytesRECV;

private:

    Protocol(const Protocol&);
    Protocol& operator = (const Protocol&);

};

