#include "CounterManager.h"

#include <windows.h>

CRITICAL_SECTION CriticalSectionCounter;

CounterManager* CounterManager::impl = NULL;

void CounterManager::initialize()
{
    InitializeCriticalSection(&CriticalSectionCounter);
    impl = new CounterManager();
}

void CounterManager::registerRequest(int size, bool success)
{
    acquire();
    impl->doRegisterRequest(size, success);
    release();
}

Counter CounterManager::get()
{
    Counter ret;
    acquire();
    ret = impl->cntr;
    release();
    return ret;
}

void CounterManager::acquire()
{
    EnterCriticalSection(&CriticalSectionCounter);
}

void CounterManager::release()
{
    LeaveCriticalSection(&CriticalSectionCounter);
}

CounterManager::CounterManager() : cntr()
{
}

CounterManager::~CounterManager()
{
}

void CounterManager::doRegisterRequest(int size, bool success)
{
    cntr.registerRequest(size, success);
}

