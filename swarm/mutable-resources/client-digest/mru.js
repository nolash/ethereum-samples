var web3 = require("web3");

if (module !== undefined) {
	module.exports = {
		digest: feedUpdateDigest
	}
}

var topicLength = 32;
var userLength = 20;
var timeLength = 7;
var levelLength = 1;
var headerLength = 8;
var updateMinLength = topicLength + userLength + timeLength + levelLength + headerLength;

function feedUpdateDigest(request /*request*/, data /*UInt8Array*/) {
	var topicBytes = undefined;
    var userBytes = undefined;
    var protocolVersion = 0;
  
    protocolVersion = request.protocolVersion

	try {
		topicBytes = web3.utils.hexToBytes(request.feed.topic);
	} catch(err) {
		console.error("topicBytes: " + err);
		return undefined;
	}

	try {
		userBytes = web3.utils.hexToBytes(request.feed.user);
	} catch(err) {
		console.error("topicBytes: " + err);
		return undefined;
	}

	var buf = new ArrayBuffer(updateMinLength + data.length);
	var view = new DataView(buf);
    var cursor = 0;
    
    view.setUint8(cursor, protocolVersion) // first byte is protocol version.
    cursor+=headerLength; // leave the next 7 bytes (padding) set to zero

	topicBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	userBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});
	
	// time is little-endian
	view.setUint32(cursor, request.epoch.time, true);
	cursor += 7;

	view.setUint8(cursor, request.epoch.level);
	cursor++;

	data.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
    });
    //console.log(web3.utils.bytesToHex(new Uint8Array(buf)))

	return web3.utils.sha3(web3.utils.bytesToHex(new Uint8Array(buf)));
}

request = {"feed":{"topic":"0x4b680e0ac418934e5468daea238ffc8e25941200ff35a99eee2b1c52357f8c2a","user":"0x1d66d3fa0250e6e3085ec4ee90a7eafb176ebfb8"},"epoch":{"time":1544905717,"level":25},"protocolVersion":0}
data = new Uint8Array([0x66, 0x6f, 0x6f])

// data payload
//data = new Uint8Array([5,154,15,165,62])

// request template, obtained calling http://localhost:8500/bzz-feed:/?user=<0xUSER>&topic=<0xTOPIC>&meta=1
//request = {"feed":{"topic":"0x1234123412341234123412341234123412341234123412341234123412341234","user":"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"},"epoch":{"time":1538650124,"level":25},"protocolVersion":0}

// obtain digest
digest = feedUpdateDigest(request, data)

console.log(digest)
