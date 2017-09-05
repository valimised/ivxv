#include "Protocol.h"
#include "Error.h"
#include "Slotter.h"
#include "CounterManager.h"

#define ENTROPY_SUCCESS ((char)0xFF)
#define ENTROPY_FAIL ((char)0x00)

void CALLBACK WorkerRoutine(DWORD error, DWORD bytesTransferred,
                            LPWSAOVERLAPPED overlapped, DWORD)
{
    Protocol* prot = (Protocol*)overlapped->hEvent;

    if (error != 0 || bytesTransferred == 0) {
        delete prot;
        return;
    }

    prot->handle(bytesTransferred);
}

Protocol::Protocol(SOCKET accept)
{
    socket = accept;
    ZeroMemory(&(overlapped), sizeof(WSAOVERLAPPED));
    bytesSEND = 0;
    bytesRECV = 0;
    bytesTOBESENT = 0;
    dataBuf.len = DATABUFSIZE;
    dataBuf.buf = buffer;
    overlapped.hEvent = this;
}

Protocol::~Protocol()
{
    closesocket(socket);
}


bool Protocol::recv()
{
    DWORD flags = 0;
    int ret = WSARecv(socket, &dataBuf, 1, NULL, &flags, &overlapped,
                      WorkerRoutine);

    if ((ret == SOCKET_ERROR) && (WSAGetLastError() != WSA_IO_PENDING)) {
        return false;
    }

    return true;
}


void Protocol::handle(DWORD bytesTransferred)
{
    int ret = 0;
    DWORD flags = 0;

    if (bytesRECV == 0) {
        bytesRECV = bytesTransferred;
        bytesSEND = 0;
        DWORD bytes = 0;
        memcpy(&bytes, dataBuf.buf, bytesTransferred);
        DWORD res = ntohl(bytes);
        bool succ = Slotter::request((BYTE*)(buffer + 1), res);
        CounterManager::registerRequest(res, succ);

        if (succ) {
            buffer[0] = ENTROPY_SUCCESS;
            bytesTOBESENT = res + 1;
        }
        else {
            buffer[0] = ENTROPY_FAIL;
            bytesTOBESENT = 1;
        }
    }
    else {
        bytesSEND += bytesTransferred;
    }

    if (bytesTOBESENT > bytesSEND) {
        ZeroMemory(&overlapped, sizeof(WSAOVERLAPPED));
        overlapped.hEvent = this;
        dataBuf.buf = buffer + bytesSEND;
        dataBuf.len = bytesTOBESENT - bytesSEND;
        ret = WSASend(socket, &dataBuf, 1, NULL, 0, &overlapped, WorkerRoutine);

        if ((ret == SOCKET_ERROR) && (WSAGetLastError() != WSA_IO_PENDING)) {
            OutputError("WSASend()", WSAGetLastError());
            return;
        }
    }
    else {
        bytesRECV = 0;
        flags = 0;
        ZeroMemory(&overlapped, sizeof(WSAOVERLAPPED));
        overlapped.hEvent = this;
        dataBuf.len = DATABUFSIZE;
        dataBuf.buf = buffer;
        ret = WSARecv(socket, &dataBuf, 1, NULL, &flags, &overlapped,
                      WorkerRoutine);

        if ((ret == SOCKET_ERROR) && (WSAGetLastError() != WSA_IO_PENDING)) {
            OutputError("WSARecv()", WSAGetLastError());
            return;
        }
    }
}

