var web3 = require("web3");

zeros = "0000000000000000000000000000000000000000000000000000000000000000";

function createManifestFromHex(topic, name, user, epoch, level) {

	// check user address param sanity
	if (!web3.utils.isHex(user)) {
		throw "user must be hex";
		return "";
	}
	if (user.substring(0, 2) != "0x") {
		user = "0x"+user;
	}
	if (user.length != 42) {
		throw "user must be 20 byte ethereum address";
		return "";
	}

	// check topic sanity
	if (!web3.utils.isHex(topic)) {
		throw "topic must be hex";
		return "";
	}

	// check name sanity
	if (typeof name !== "string") {
		throw "name must be string";
		return "";
	} else if (name.length > 32) {
		throw "name must be maximum 32 bytes long";
		return "";
	}

	// topic should be a 32 byte padded hex string in the manifest
	var topicPadded = zeros + topic;
	topic = topicPadded.slice(-64);


	var topicXored = "";
	var nameLimit;
	if (name.length > 32) {
		nameLimit = 32;
	} else {
		nameLimit = name.length;	
	}
	for (i = 0; i < topic.length / 2; i++) {
		var nameByte = "";
		if (i < name.length) {
			nameByte = name.charCodeAt(i);
		}
		var topicByte = parseInt(topic.substring(i*2, (i*2)+2), 16);
		topicXored += ("0"+(nameByte ^ topicByte).toString(16)).slice(-2);
		//console.log("name " + nameByte + " topic " + topicByte + " = " + (nameByte ^ topicByte));
	}

	return '{"entries":[{"contentType":"application/bzz-feed","mod_time":"0001-01-01T00:00:00Z","feed":{"topic":"0x' + topicXored + '","user":"' + user + '"}}]}';

}

console.log(createManifestFromHex("0x660000000000000000000000000000000000000000000000000000000000002a", "foo", "19cb96e2fcf9afd95ef06a504ca4feb89c05ca88"));
