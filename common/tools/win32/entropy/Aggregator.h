#pragma once

#include <windows.h>
#include <wincrypt.h>


typedef struct {
    BYTE bScanCode;
    DWORD dwTickCount;
    BOOL isUp;
} KeyPressed;


typedef struct {
    POINT ptMousePos;
    DWORD dwTickCount;
} MousePosition;


class Aggregator {

public:

    Aggregator();
    virtual ~Aggregator();

    BOOL init();
    BOOL enoughEntropy();
    BOOL prepareSlice();
    BYTE* getSlice();

    double entropy() const;
    DWORD requested() const;
    BOOL handle(MousePosition* event);
    BOOL handle(KeyPressed* event);

    static BOOL initialize();
    static int outbytes();
    static void finalize();

protected:

    double dEntropy;
    DWORD cbRequested;
    POINT ptLastPos;
    BYTE bLastScanCode;
    BYTE* pbHashData;
    DWORD dwLastTime;
    HCRYPTHASH hHash;

private:

    Aggregator(const Aggregator&);
    Aggregator& operator = (const Aggregator&);

    static HCRYPTPROV hProvider;
};

