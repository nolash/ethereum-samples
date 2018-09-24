var web3 = require("web3");
var ethtx = require("ethereumjs-tx");

if (module !== undefined) {
	module.exports = {
		digest: mruDigest
	}
}

var topicLength = 32;
var userLength = 20;
var timeLength = 7;
var levelLength = 1;
var updateMinLength = topicLength + userLength + timeLength + levelLength;

function mruDigest(o) {
	var topicBytes = undefined;
	var dataBytes = undefined;
	var userBytes = undefined;
	
	if (!web3.utils.isHexStrict(o.data)) {
		console.error("data must be a valid 0x prefixed hex value");
		return undefined;
	}

	dataBytes = web3.utils.hexToBytes(o.data);

	try {
		topicBytes = web3.utils.hexToBytes(o.topic);
	} catch(err) {
		console.error("topicBytes: " + err);
		return undefined;
	}

	try {
		userBytes = web3.utils.hexToBytes(o.user);
	} catch(err) {
		console.error("topicBytes: " + err);
		return undefined;
	}

	var buf = new ArrayBuffer(updateMinLength + dataBytes.length);
	var view = new DataView(buf);
	var cursor = 0;

	topicBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	userBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});
	
	// time is little-endian
	view.setUint32(cursor, o.time, true);
	cursor += 7;

	view.setUint8(cursor, o.level);
	cursor++;

	dataBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	return web3.utils.sha3(web3.utils.bytesToHex(new Uint8Array(buf)));
}

