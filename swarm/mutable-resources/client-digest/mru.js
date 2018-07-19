var web3 = require("web3");

var bzzKeyLength = 32;
var mruMetaHashLength = bzzKeyLength;
var mruRootAddrLength = bzzKeyLength;
var mruUpdateVersionLength = 4;
var mruUpdatePeriodLength = 4;
var mruUpdateFlagLength = 1;
var mruUpdateDataLengthLength = 2;
var mruUpdateHeaderLengthLength = 2;
var mruUpdateHeaderLength = mruUpdateFlagLength + mruUpdatePeriodLength + mruUpdateVersionLength + mruMetaHashLength + mruRootAddrLength;
var mruUpdateMinLength = mruUpdateHeaderLength + mruUpdateDataLengthLength + mruUpdateHeaderLengthLength;

var _buf = new ArrayBuffer(mruUpdateHeaderLengthLength);
var _view = new DataView(_buf);
_view.setUint16(0, mruUpdateHeaderLength);
var mruUpdateHeaderLengthBytes = new Uint8Array(_buf);
delete _buf;
delete _view;

function mruUpdateDigest(o) {

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

console.log(mruUpdateDigest({
	"period": 42,
	"version": 13,
	"multihash": false,
	"data": "0x666f6f",
	"metaHash": "0x2c1183eed6a4b0046da699e2655a406d20754ef02fcc7625ee24579a4c0970eb", 
	"rootAddr": "0xfe9a53da332939697dd3b2d706f161ba75162805752efe7d365f2ed3f5cbd380" 
}));
