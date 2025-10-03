#!/bin/bash

TOTAL=$(($1-1))

source card_options.conf

function erase {
    # erase-card will explicitly ignore any PIN from commandline/env, must always enter manually
    echo $PKCS15_SO_PIN | pkcs15-init --erase-card -r $1 ;
}

echo "During erasure, the pkcs15-init script asks for Security Officer PIN."
echo "This is provided automagically, you don't need to insert anything".
echo "Please wait until erasing and initalizing $1 smart cards"
echo "==============================="

for i in `seq 0 $TOTAL`; do
  echo "Erasing card $i"
    erase $i;
  echo "Done"
  echo
done

echo "==============================="
echo "All done."
