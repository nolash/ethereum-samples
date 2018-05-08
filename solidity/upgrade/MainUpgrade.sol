pragma solidity ^0.4.0;

contract MainUpgrade is AbstractMain {
	uint16 public plugh;
	function MainUpgrade() {
		plugh = 666;
	}
}
