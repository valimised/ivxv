#!/bin/bash

TOTAL=$(($1-1))
SOPIN=`grep so-pin card_options.conf | cut -d" " -f2`
PIN=`grep pin card_options.conf | grep -v so | cut -d" " -f2`

function erase {
    echo $SOPIN | pkcs15-init --erase-card -r $1 ;
}

function create {
    pkcs15-init --options-file card_options.conf --create-pkcs15 -r $1;
    pkcs15-init --options-file card_options.conf --auth-id 01 --store-pin --label el -r $1;
    pkcs15-init --finalize -r $1;

}

echo "PIN will be $PIN."
echo "During initalization, the pkcs15-init script asks for Security Officer PIN."
echo "This is provided automagically, you don't need to insert anything".
echo "Please wait until erasing and initalizing $1 smart cards"
echo "==============================="

for i in `seq 0 $TOTAL`; do
    erase $i;
    create $i;
done

echo "==============================="
echo "All done."
echo "PIN is $PIN."
