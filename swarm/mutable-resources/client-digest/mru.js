var web3 = require("web3");

var topicLength = 32;
var userLength = 20;
var timeLength = 7;
var levelLength = 1;
var updateMinLength = topicLength + userLength + timeLength + levelLength;

function mruUpdateDigest(o) {
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
	
	// time is little endian
	var timeBuf = new ArrayBuffer(4);
	var timeView = new DataView(timeBuf);
	//view.setUint32(cursor, o.time);
	timeView.setUint32(0, o.time);
	var timeBufArray = new Uint8Array(timeBuf);
	for (i = 0; i < 4; i++) {
		view.setUint8(cursor, timeBufArray[3-i]);
		cursor++;
	}

	for (i = 0; i < 3; i++) {
		view.setUint8(cursor, 0);
		cursor++;
	}

	//cursor += 4;
	view.setUint8(cursor, o.level);
	cursor++;

	dataBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	return web3.utils.sha3(web3.utils.bytesToHex(new Uint8Array(buf)));
}

mru = mruUpdateDigest(
	{	
		"topic": "0x0102000000000000000000000000000000000000000000000000000000000304",
		"data": "0x666f6f",	
		"user": "0x44defd3d4e99c2f97e68f3a0a6462e590fd10b91",
		"time": 1536427515,
		"level": 25,
	}
);
	
console.log(mru);
