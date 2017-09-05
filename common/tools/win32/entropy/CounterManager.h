#pragma once

#include "Counter.h"

class CounterManager {

public:

    static void initialize();
    static void registerRequest(int size, bool success);
    static Counter get();

protected:

    static void acquire();
    static void release();

    CounterManager();
    virtual ~CounterManager();

    void doRegisterRequest(int size, bool success);

    Counter cntr;

private:

    static CounterManager* impl;
    CounterManager(const CounterManager&);
    CounterManager& operator = (const CounterManager&);

};

