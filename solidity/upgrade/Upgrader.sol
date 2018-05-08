pragma solidity ^0.4.0; 

import './Store.sol';
import './AbstractMain.sol';

contract Upgrader {
	/**
	 * holds the address of the latest version of Main
	 * servers as authenticator for those wishing to use the Main contract
	 */
	address public current;

	event Upgraded(address newaddress);

	/**
	 * enable contract to receive funds on creation 
	 */
	function Upgrader() payable {
	}

	/**
	 * creates new main and store instances
	 * assigns new store to main contract
	 * transfers eth from upgrader contract to main
	 */
	function create() {
		Store store = new Store();
		store.poke(); // store foo set to true
		AbstractMain main =  new AbstractMain();
		main.setStore(store);
		main.transfer(this.balance);
		current = main;
	}

	/**
	 * accepts a main instance, and assigns the existing store of the current Main to it
	 * transfers eth from the current Main to the new Main contract
	 * the Upgrader contract MUST be the owner of the new Main contract (see AbstractMain.transfer(address))
	 * @param _main address of new Main instance
	 */
	function upgrade(address _main) {
		AbstractMain newmain = AbstractMain(_main);
		require(newmain.isOwned()); // will throw if AbstractMain.transfer with our address hasn't been called
		AbstractMain oldmain = AbstractMain(current);
		Store store = oldmain.store();
		newmain.setStore(store); // will throw if store is already set
		oldmain.kill(newmain); 
		current = newmain; // we know we own the new contract, and moneys have been passed
		Upgraded(newmain);
	}
	
	/**
	 * for verifying that the Store instance has persisted
	 * and for verifying that eth has been transferred to new Main after upgrade/creation
	 */
	function check() constant returns(bool, uint256) {
	    AbstractMain main = AbstractMain(current);
	    return (main.store().foo(), main.balance);
	}
}
