# Upgradeable contract

An attempt at a simple example of how to implement a contract where the bytecode of the main interface can be changed while retaining persistent storage and eth balance.

The `Upgrader.create()` will only create an instance of the abstract contract. The intention is that `Upgrader.upgrade(address)` be called immediately after creation, passing the first proper implementation of the contract.

## Running in remix

Run `sh remix.sh > remixcodefile.txt` and paste the contents of `remixcodefile.txt` into remix.

First we instantiate the upgrader and create the first contract:

1. Set value > 0
2. Upgrader -> Create
3. Upgrader -> create
4. Upgrader -> current: != 0x0
5. Upgrader -> check: bool == true, uint256 > 0

Attempt upgrade, first without and with transfer of ownership of Main contract:

1. Main -> Create
2. copy address of Main
3. Upgrader -> upgrade("address of Main"): throws
4. copy address of Upgrader
5. Main -> transfer("address of Upgrader")
6. Upgrader -> upgrade("address of Main"): triggers event Upgrade with address of Main
7. Upgrader -> current(): == address of Main
8. Upgrader -> check: bool == true, uint256 > 0

Attempt to alter store after upgrade

1. Store -> Create
2. copy address of Store
3. Main -> setStore("address of Store"): throws

Attempt another upgrade, with preset store in Main contract

1. Store -> Create
2. copy new address of Store
3. MainUpgrade -> Create
4. MainUpgrade -> setStore("address of Store")
5. copy address of MainUpgrade
6. Upgrader -> upgrade("address of MainUpgrade"): throws 
7. Upgrader -> current(): == address of Main (not MainUpgrade)

Attempt to steal eth by killing contract explicitly, transferring ownership

1. copy address of MainUpgrade
2. Main -> kill("address of MainUpgrade"): throws because Main is owned by Upgrader
3. Main -> transfer("address of MainUpgrade"): silently doesn't execute
4. Main -> kill("address of MainUpgrade"): throws because Main is owned by Upgrader
5. Main -> check(): bool == true, uint256 > 0
