pragma solidity ^0.4.20;

contract escrow {

	uint256 public seq;

    // statuses:
    // 0 = not active
    // 1 = active
    // 2 = paid 
    // 3 = released
    // 4 = withdrawn
    // 5 = cancelled
	struct participant {
		uint80 amount;
		uint8 status;
		address beneficiary;
	}

	struct escrowItem {
		bytes8 name;
		address owner;
	    mapping(address => participant) participants;
	    address[] participantsIdx;
	}

	participant[] participantsRegistry;
	escrowItem[] escrowItems;

	event Created(uint256 seq, bytes8 _name);
    event ParticipantPaid(uint256 seq, address participant);
    event ParticipantReleased(uint256 seq, address participant);
    event ParticipantCancelled(uint256 seq, address participant);
    event Finished(uint256 seq);
    

	function createEscrow(bytes8 _name) public {
		escrowItem memory e;
		seq++;
		
		e.owner = msg.sender;
		e.name = _name;
	    	escrowItems.push(e);
		emit Created(seq, _name);
	}
	
	function addParticipant(uint256 _seq, uint80 _amount, address _who, address _beneficiary) public {
		require(msg.sender == escrowItems[_seq-1].owner);
		require(escrowItems[_seq-1].participants[_who].status == 0);

		participant memory p;
		
		p.amount = _amount;
		p.beneficiary = _beneficiary;
		p.status = 1;
		
		escrowItems[_seq-1].participants[_who] = p;
		escrowItems[_seq-1].participantsIdx.push(_who);
	}
	
	// add an alias function without the tip steps
	function participantPut(uint256 _id) public payable {
		require(escrowItems[_id].participants[msg.sender].beneficiary != address(0x0));
		require(escrowItems[_id].participants[msg.sender].amount <= msg.value);

		escrowItems[_id].participants[msg.sender].amount = uint80(msg.value);
		escrowItems[_id].participants[msg.sender].status = 2;
		emit ParticipantPaid(_id, msg.sender);
	}
	
	function participantRelease(uint256 _id) public {
	    require(escrowItems[_id].participants[msg.sender].status == 2);
        
        escrowItems[_id].participants[msg.sender].status = 3;
        emit ParticipantReleased(_id, msg.sender);
	}
	
	// can only withdraw if all have released or if cancelled
	function participantPayout(uint256 _id) public returns (bool) {
	    require(escrowItems[_id].participants[msg.sender].status > 2);
	    
	    // mode is 3 for release and 5 for cancelled
	    // for payout to succeed, all must be released or own must be cancelled
	    uint8 mode = 3;
	    for (uint8 i = 0; i < escrowItems[_id].participantsIdx.length; i++) {
	        if (escrowItems[_id].participants[escrowItems[_id].participantsIdx[i]].status == 5) {
	            mode = 5;
	            break;
	        }
	        require(escrowItems[_id].participants[escrowItems[_id].participantsIdx[i]].status == 3);
	    }
	    
	    uint256 amountToSend = escrowItems[_id].participants[msg.sender].amount;
	    
	    // if one is cancelled but not self, require cancel first
	    // else if cancelled, return to self
	    // else send to beneficiary
	    address beneficiary;
	    
	    if (escrowItems[_id].participants[msg.sender].status == 5) {
	        beneficiary = msg.sender;
	    } else if (mode == 5) {
	        revert();
	    } else {
	        beneficiary = escrowItems[_id].participants[msg.sender].beneficiary;
	    } 
	    escrowItems[_id].participants[msg.sender].status = 4;
	    escrowItems[_id].participants[msg.sender].amount = 0;
	    if (!msg.sender.send(escrowItems[_id].participants[msg.sender].amount)) {
            escrowItems[_id].participants[msg.sender].amount = uint80(amountToSend);
            escrowItems[_id].participants[msg.sender].status = mode;
            return false;
        }
        return true;
	}
	
	// can cancel at any time except:
	// if payout has been made
	// if all participants including sender have released
	function participantCancel(uint256 _id) public {
	    require(escrowItems[_id].participants[msg.sender].status < 4);
	    
	    uint8 releases;
	    for (uint8 i = 0; i < escrowItems[_id].participantsIdx.length; i++) {
	        if (escrowItems[_id].participants[escrowItems[_id].participantsIdx[i]].status == 3) {
	            releases++;
	        }
	    }
	    require(releases != escrowItems[_id].participantsIdx.length);
	    escrowItems[_id].participants[msg.sender].status = 5;
	    emit ParticipantCancelled(_id, msg.sender);
	}
}

