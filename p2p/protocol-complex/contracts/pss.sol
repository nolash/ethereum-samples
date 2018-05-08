pragma solidity ^0.4.21;

contract pss {
	mapping (address => uint256) deposits;
	mapping (address => uint256) credits;

	// custom functions
	function deposit() public payable {
		deposits[msg.sender] += msg.value;
	}

	function transfer(uint256 amount, address beneficiary) public {
		require(deposits[msg.sender] >= amount);
		deposits[msg.sender] -= amount;
		credits[beneficiary] += amount;
	}

	function withdraw(uint256 amount) public {
		require(credits[msg.sender] >= amount);

		credits[msg.sender] -= amount;
		if (!msg.sender.send(amount)) {
			credits[msg.sender] += amount;
		}
	}
}
