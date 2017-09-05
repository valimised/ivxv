#include "Aggregator.h"

#define OUTBYTES 20
#define OUTBITS (OUTBYTES * 8)
#define MOUSE_ENTROPY_PER_SAMPLE 1.5
#define KEY_ENTROPY_PER_SAMPLE 1

HCRYPTPROV Aggregator::hProvider = 0;

BOOL Aggregator::initialize()
{
    return CryptAcquireContext(&hProvider, 0, MS_DEF_PROV, PROV_RSA_FULL,
                               CRYPT_VERIFYCONTEXT);
}

int Aggregator::outbytes()
{
    return OUTBYTES;
}

void Aggregator::finalize()
{
    if (hProvider) {
        CryptReleaseContext(hProvider, 0);
    }
}


Aggregator::Aggregator()
{
    dEntropy = 0.0;
    cbRequested = OUTBITS;
    ptLastPos.x = 0;
    ptLastPos.y = 0;
    bLastScanCode = 0;
    pbHashData = NULL;
    dwLastTime = 0;
    hHash = 0;
}

DWORD Aggregator::requested() const
{
    return cbRequested;
}

double Aggregator::entropy() const
{
    return dEntropy;
}

BOOL Aggregator::init()
{
    BOOL ret = false;
    DWORD dwByteCount = sizeof(DWORD);
    DWORD cbHashData = 0;
    HCRYPTHASH tmpHash = 0;
    BYTE* tmpHashData = 0;

    if (!GetCursorPos(&ptLastPos)) {
        goto error;
    }

    if (!CryptCreateHash(hProvider, CALG_SHA1, 0, 0, &tmpHash)) {
        goto error;
    }

    if (!CryptGetHashParam(tmpHash, HP_HASHSIZE, (BYTE*)&cbHashData,
                           &dwByteCount, 0)) {
        goto error;
    }

    if (OUTBYTES > cbHashData) {
        goto error;
    }

    tmpHashData = (BYTE*)LocalAlloc(LMEM_FIXED, cbHashData);

    if (tmpHashData == NULL) {
        goto error;
    }

    hHash = tmpHash;
    tmpHash = 0;
    pbHashData = tmpHashData;
    tmpHashData = 0;
    ret = true;
error:

    if (tmpHashData) {
        LocalFree(tmpHashData);
    }

    if (tmpHash) {
        CryptDestroyHash(tmpHash);
    }

    return ret;
}


Aggregator::~Aggregator()
{
    if (pbHashData) {
        LocalFree(pbHashData);
        pbHashData = NULL;
    }

    if (hHash) {
        CryptDestroyHash(hHash);
        hHash = 0;
    }
}

BOOL Aggregator::enoughEntropy()
{
    return (dEntropy >= cbRequested);
}

BOOL Aggregator::prepareSlice()
{
    return CryptGetHashParam(hHash, HP_HASHVAL, pbHashData, &cbRequested, 0);
}

BYTE* Aggregator::getSlice()
{
    BYTE* ret = pbHashData;
    pbHashData = NULL;
    return ret;
}

BOOL Aggregator::handle(MousePosition* event)
{
    //TODO! Err not handled
    CryptHashData(hHash, (BYTE*)event, sizeof(*event), 0);

    if ((event->ptMousePos.x != ptLastPos.x || event->ptMousePos.y !=
            ptLastPos.y) && event->dwTickCount - dwLastTime > 100) {
        ptLastPos = event->ptMousePos;
        dwLastTime = event->dwTickCount;
        dEntropy += MOUSE_ENTROPY_PER_SAMPLE;
        return TRUE;
    }

    return FALSE;
}

BOOL Aggregator::handle(KeyPressed* event)
{
    // TODO! Err not handled
    CryptHashData(hHash, (BYTE*)event, sizeof(*event), 0);

    if (event->isUp || (bLastScanCode != event->bScanCode && event->dwTickCount
                        - dwLastTime > 100)) {
        bLastScanCode = event->bScanCode;
        dwLastTime = event->dwTickCount;
        dEntropy += KEY_ENTROPY_PER_SAMPLE;
        return TRUE;
    }

    return FALSE;
}

