pragma solidity ^0.4.0;

contract Store {
	bool public foo;

	/**
	 * for proof of persistent store
	 * set this explicitly when the first contract is created
	 */
	function poke() {
		foo = true;
	}
}
