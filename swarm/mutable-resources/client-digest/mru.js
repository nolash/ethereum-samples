var web3 = require("web3");

//var bzzKeyLength = 32;
//var mruMetaHashLength = bzzKeyLength;
//var mruRootAddrLength = bzzKeyLength;
//var mruUpdateVersionLength = 4;
//var mruUpdatePeriodLength = 4;
//var mruUpdateFlagLength = 1;
//var mruUpdateDataLengthLength = 2;
//var mruUpdateHeaderLengthLength = 2;
//var mruUpdateHeaderLength = mruUpdateFlagLength + mruUpdatePeriodLength + mruUpdateVersionLength + mruMetaHashLength + mruRootAddrLength;
//var mruUpdateMinLength = mruUpdateHeaderLength + mruUpdateDataLengthLength + mruUpdateHeaderLengthLength;

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
	
	for (i = 0; i < 3; i++) {
		view.setUint8(cursor, 0);
		cursor++;
	}

	view.setUint32(cursor, o.time);
	cursor += 4;
	view.setUint8(cursor, o.level);
	cursor++;

	dataBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	return web3.utils.bytesToHex(new Uint8Array(buf));
}

function mruUpdateDigest_(o) {

	var metaHashBytes = undefined;
	var rootAddrBytes = undefined;
	var dataBytes = web3.utils.hexToBytes(o.data);

	if (!web3.utils.isHexStrict(o.data)) {
		console.error("data must be a valid 0x prefixed hex value");
		return undefined;
	}

	try {
		metaHashBytes = web3.utils.hexToBytes(o.metaHash);
	} catch(err) {	
		console.error("metaHash: " + err);
		return undefined;
	}
	if (metaHashBytes.length < mruMetaHashLength) {
		console.error("metaHash must be exactly " + mruMetaHashLength + " bytes long");
		return undefined;
	}

	try {
		rootAddrBytes = web3.utils.hexToBytes(o.rootAddr)
	} catch(err) {
		console.error("rootAddr: " + err);
		return undefined;
	}
	if (rootAddrBytes.length < mruRootAddrLength) {
		console.error("rootAddr must be exactly " + mruRootAddrLength + " bytes long");
		return undefined;
	}

	var buf = new ArrayBuffer(mruUpdateMinLength + dataBytes.length);
	var view = new DataView(buf);
	var cursor = 0;
	
	view.setUint16(cursor, mruUpdateHeaderLength, true);
	cursor += mruUpdateHeaderLengthLength;
	
	view.setUint16(cursor, dataBytes.length, true);
	cursor += mruUpdateDataLengthLength;
	
	view.setUint32(cursor, o.period, true);
	cursor += mruUpdatePeriodLength;

	view.setUint32(cursor, o.version, true);
	cursor += mruUpdateVersionLength;

	rootAddrBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	metaHashBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;
	});

	if (o.multihash) {
		view.setUint8(cursor, 1);
	} else {
		view.setUint8(cursor, 0);
	}
	cursor++;

	dataBytes.forEach(function(v) {
		view.setUint8(cursor, v);
		cursor++;	
	});

	return web3.utils.sha3(web3.utils.bytesToHex(new Uint8Array(buf)));
}

//mru = mruUpdateDigest(
//{
//  "rootAddr": "0x1c2692b93e594a47af49f3cff10e2c958e00d8e5014b37e4150809f789ccd16a",
//  "metaHash": "0xa5e0aa94cce20457061e35a3b0b698e6e2cfb82b1e7ea211d9d92284f3d18b17",
//  "version": 1,
//  "period": 2,
//  "multihash": true,
//  "data": "0x1b20d7d24499dcff78f02db5291fca0f4cbbcbd691ad15d67268d0cb14ee3fc9fb86",
//});

mru = mruUpdateDigest(
	{	
		"topic": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"data": "0x666f6f",	
		"user": "0x44defd3d4e99c2f97e68f3a0a6462e590fd10b91",
		"time": 1536410917,
		"level": 0
	}
);
	
console.log(mru);
