var web3 = require("web3");

var bzzKeyLength = 32;
var mruMetaHashLength = bzzKeyLength;
var mruRootAddrLength = bzzKeyLength;
var mruUpdateVersionLength = 4;
var mruUpdatePeriodLength = 4;
var mruUpdateFlagLength = 1;
var mruUpdateDataLengthLength = 2;
var mruUpdateHeaderLengthLength = 2;
var mruUpdateHeaderLength = mruUpdateFlagLength + mruUpdatePeriodLength + mruUpdateVersionLength + mruMetaHashLength + bzzKeyLength;

function unsignedNumberToBytesLE(n, l) {
	if (isNaN(parseInt(n, 10))) {
		throw "invalid number";
	}
	var buf = new ArrayBuffer(l);
	var v = new DataView(buf);
	switch(l) {
		case 1:
			v.setUint8(0, n, true);
			break;
		case 2:
			v.setUint16(0, n, true);
			break;
		case 4:
			v.setUint32(0, n, true);
			break;
		default:
			return undefined;
	}
						
	return Array.from(new Uint8Array(buf));
}

var mruUpdateHeaderLengthBytes = unsignedNumberToBytesLE(mruUpdateHeaderLength, mruUpdateHeaderLengthLength);

function mruUpdateDigest(o) {

	var metaHashBytes = undefined;
	var rootAddrBytes = undefined;
	var dataBytes = web3.utils.hexToBytes(o.data);
	var b = Array.from(mruUpdateHeaderLengthBytes);

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

	unsignedNumberToBytesLE(dataBytes.length, mruUpdateDataLengthLength).forEach(function(v) {
		b.push(v);
	});

	try {
		unsignedNumberToBytesLE(o.period, mruUpdatePeriodLength).forEach(function(v) {
			b.push(v);
		});
	} catch(e) {
		console.log("period: " + e);
		return undefined;
	}

	try {
		unsignedNumberToBytesLE(o.version, mruUpdateVersionLength).forEach(function(v) {
			b.push(v);
		});
	} catch(e) {
		console.log("version: " + e);
		return undefined;
	}

	rootAddrBytes.forEach(function(v) {
		b.push(v);
	});

	metaHashBytes.forEach(function(v) {
		b.push(v);
	});

	b.push(o.multihash ? 0x01 : 0x00);

	dataBytes.forEach(function(v) {
		b.push(v);
	});

	return web3.utils.sha3(web3.utils.bytesToHex(b));
}

console.log(mruUpdateDigest({
	"period": 42,
	"version": 13,
	"multihash": false,
	"data": "0x666f6f",
	"metaHash": "0x2c1183eed6a4b0046da699e2655a406d20754ef02fcc7625ee24579a4c0970eb", 
	"rootAddr": "0xfe9a53da332939697dd3b2d706f161ba75162805752efe7d365f2ed3f5cbd380" 
}));
