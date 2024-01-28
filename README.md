# Δ DeltaDiode
DeltaDiode, a data diode system, consists of three nodes organized in a one-way circular layout. 

This architecture ensures secure acknowledgment of data transmissions from the unsecured side to the secure side by initially hashing the acknowledgment on the secure side, and then applying an additional hashing step at the relay node when the acknowledgment is transmitted back to unsecure side.

For the DeltaDiode configuration to be compromised, both the receiver and the relay nodes, along with the return path to the transmitter, must be breached.
