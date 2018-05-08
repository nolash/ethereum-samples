pragma solidity ^0.4.0;

import './Store.sol';
import './lib/owned.sol';

contract AbstractMain is owned {
	Store public store;

	/**
	 * set the data store for the contract
	 * can only be called if a store doesn't yet exist
	 * that way we can't cheat and change the store after upgrade
	 * @param _store address of persistent store contract
	 */
	function setStore(address _store) {
		address empty;
		require(store == empty);
		store = Store(_store);
	}

	/**
	 * deletes contract, and transfers funds to new contract
	 * only owner can do this
	 * @param _beneficiary address of contract to receive funds
	 */
	function kill(address _beneficiary) {
		require(isOwned());
		selfdestruct(_beneficiary);
	}

	/**
	 * transfers ownership of contract to new owner
	 * @param _owner address of new owner
	 */
	function transfer(address _owner) onlyowner {
		owner = _owner;
	}

	/**
	 * returns true if owner of contract is current called
	 */
	function isOwned() constant returns(bool) {
		return owner == msg.sender;
	}

	/**
	 * fallback for accepting payments
	 */
	function () payable {}
}
