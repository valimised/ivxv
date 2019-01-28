#include <windows.h>
#include <winsock2.h>

#include "Error.h"
#include "Protocol.h"

#define PORT 22062

SOCKET acceptSocket;

DWORD WINAPI WorkerThread(LPVOID lpParameter)
{
    DWORD ret = 1;
    Protocol* protHandler = NULL;
    DWORD i = 0;
    WSAEVENT events[1];
    events[0] = (WSAEVENT) lpParameter;

    while (true) {
        while (true) {
            i = WSAWaitForMultipleEvents(1, events, FALSE, WSA_INFINITE, TRUE);

            if (i == WSA_WAIT_FAILED) {
                OutputError("WSAWaitForMultipleEvents()", WSAGetLastError());
                goto error;
            }

            if (i != WAIT_IO_COMPLETION) {
                break;
            }
        }

        if (WSAResetEvent(events[i - WSA_WAIT_EVENT_0]) == FALSE) {
            OutputError("WSAWaitForMultipleEvents()", WSAGetLastError());
            goto error;
        }

        protHandler = new Protocol(acceptSocket);

        if (protHandler == NULL) {
            OutputError("GlobalAlloc()", GetLastError());
            goto error;
        }

        if (!protHandler->recv()) {
            OutputError("WSARecv()", WSAGetLastError());
            goto error;
        }
    }

    // never reached
    ret = 0;
error:
    return ret;
}


DWORD WINAPI AcceptThread(LPVOID)
{
    WSADATA wsaData;
    SOCKADDR_IN internetAddr;
    HANDLE threadHandle = NULL;
    DWORD threadId = 0;
    SOCKET listenSocket = 0;
    WSAEVENT acceptEvent;
    INT ret = WSAStartup(MAKEWORD(2, 2), &wsaData);

    if (ret != 0) {
        OutputError("WSAStartup()", ret);
        WSACleanup();
        return 1;
    }

    listenSocket = WSASocket(AF_INET, SOCK_STREAM, 0, NULL, 0,
                             WSA_FLAG_OVERLAPPED);

    if (listenSocket == INVALID_SOCKET) {
        OutputError("WSASocket()", WSAGetLastError());
        return 1;
    }

    internetAddr.sin_family = AF_INET;
    internetAddr.sin_addr.s_addr = htonl(INADDR_ANY);
    internetAddr.sin_port = htons(PORT);
    ret = bind(listenSocket, (PSOCKADDR) &internetAddr, sizeof(internetAddr));

    if (ret == SOCKET_ERROR) {
        OutputError("bind()", WSAGetLastError());
        return 1;
    }

    if (listen(listenSocket, 5)) {
        OutputError("listen()", WSAGetLastError());
        return 1;
    }

    acceptEvent = WSACreateEvent();

    if (acceptEvent == WSA_INVALID_EVENT) {
        OutputError("WSACreateEvent()", WSAGetLastError());
        return 1;
    }

    threadHandle = CreateThread(NULL, 0, WorkerThread, (LPVOID) acceptEvent, 0,
                                &threadId);

    if (threadHandle == NULL) {
        OutputError("CreateThread()", GetLastError());
        return 1;
    }

    while (true) {
        acceptSocket = accept(listenSocket, NULL, NULL);
        struct linger so_linger;
        so_linger.l_onoff = 1;
        so_linger.l_linger = 0;
        ret = setsockopt(acceptSocket, SOL_SOCKET, SO_LINGER, (char*)&so_linger,
                         sizeof so_linger);

        if (ret) {
            OutputError("setsockopt()", WSAGetLastError());
            return 1;
        }

        if (WSASetEvent(acceptEvent) == FALSE) {
            OutputError("WSASetEvent()", WSAGetLastError());
            return 1;
        }
    }
}

