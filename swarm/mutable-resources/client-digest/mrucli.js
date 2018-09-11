var web3 = require("web3");
var flags = require("flags");
var fs = require("fs");
var mru = require("./mru.js");

flags.defineString("topic", "", "topic (optional, 0x prefixed hex)");
flags.defineBoolean("hex", false, "data input is hex encoded");
flags.defineString("user", "", "user (0x prefixed hex)");
flags.defineInteger("time", 0, "epoch time");
flags.defineInteger("level", 25, "epoch level");

var restArgs = flags.parse();
var inBuf = undefined;

if (restArgs.length == 0) {
	inBuf = fs.readFileSync(0);
} else {
	inBuf = fs.readFileSync(restArgs[0]);
}

if (flags.get("hex")) {
	inBuf = inBuf.toString("ascii");
} else {
	inBuf = web3.utils.bytesToHex(inBuf);
}

mruData = {	
	"topic": flags.get("topic"),
	"user": flags.get("user"),
	"time": flags.get("time"),
	"level": flags.get("level"),
	"data": inBuf
};

mruResult = mru.digest(mruData);
console.log(mruResult);

