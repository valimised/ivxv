"""
Reference PseudoRandom Number Generator implementation.
=======================================================

Pseudocode of PRNG
------------------

PRNG(seed, amount): # amount is in bytes
    1. rounds = (amount + 31)/32
    2. for (i = 1; i <= rounds; i++)
    3.     out = out || H(i || seed) # counter i is 64-bit big-endian
                                     # unsigned integer
    4. end for
    5. return amount bytes of out
"""

import hashlib
import struct

DIGEST_ALGORITHM_IDENTIFIER = "SHA256"
# the counter is 64 byte big-endian unsigned integer
COUNTER_PACKING_FORMAT = ">Q"
SHOW_DEBUG = False


def D(msg, *args, **kwargs):
    if SHOW_DEBUG:
        print(msg.format(*args, **kwargs))


class PRNG(object):
    """
    >>> s = '000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f'
    >>> PRNG(bytes.fromhex(s)).read(64).hex()
    'fab69ae5a169653b318a6e2e3927fb96f6e79fec0d6e20b95de273c02b7f02002620cd8765179d8b962d51032a7df60561cb492849ce4369c7763033a09fdf0f'
    >>> s = 'ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff'
    >>> PRNG(bytes.fromhex(s)).read(64).hex()
    '169a8da0c474a25309f5934928d4f1c582e4927d7191fe90dde01a84b2f9af9b7ff3645d18399c432983b1ace0f8d6088fca43cf6289ac29a480a1fca1d7a735'
    """

    def __init__(self, seed):
        self.seed = seed
        self.counter = 1
        self.buffer = b""

        D("initialized PRNG\nseed\t'{}'", seed.hex())

    def digest_size(self):
        return self.get_digest_instance().digest_size

    def get_digest_instance(self):
        return hashlib.new(DIGEST_ALGORITHM_IDENTIFIER)

    def refill_buffer(self):
        if self.buffer:
            return
        round_seed = self.round_input()
        digest_instance = self.get_digest_instance()
        digest_instance.update(round_seed)
        self.buffer = digest_instance.digest()
        self.counter += 1

        D("updating buffer\nround seed\t'{}'\nfilled buffer\t'{}'",
          round_seed.hex(),
          self.buffer.hex())

    def read_bytes_from_buffer(self, amount_bytes):
        return_value = self.buffer[:amount_bytes]
        self.buffer = self.buffer[amount_bytes:]
        return return_value

    def read(self, amount_bytes):
        D("reading from buffer\namount\t'{}'\ncurrent buffer\t'{}'",
          amount_bytes, self.buffer.hex())

        return_value = b""
        while len(return_value) < amount_bytes:
            to_read = amount_bytes - len(return_value)
            self.refill_buffer()
            return_value += self.read_bytes_from_buffer(to_read)

        D("read from buffer\nremaining buffer\t'{}'\nreturn value\t'{}'",
          self.buffer.hex(),
          return_value.hex())

        return return_value

    def round_counter(self):
        return struct.pack(
            COUNTER_PACKING_FORMAT,
            self.counter)

    def round_input(self):
        return self.round_counter() + self.seed


if __name__ == '__main__':
    import doctest
    doctest.testmod()
