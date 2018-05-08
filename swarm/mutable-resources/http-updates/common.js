function setError(errstr, ok) {
	var color;
	switch (ok) {
		case 1, 2:
			color = "red";
			break;
		default:
			color = "green";

	}
	document.getElementById("error").innerHTML = '<font color="' + color + '">' + errstr + '</font>';
}
