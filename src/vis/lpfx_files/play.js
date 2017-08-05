// function sleep(ms) {
//   return new Promise(resolve => setTimeout(resolve, ms));
// }

function sleepFor( sleepDuration ){
    var now = new Date().getTime()
    while (new Date().getTime() < now + sleepDuration) {}
}

function visualize(lines) {
    map = JSON.parse(lines[0])
    initCy(map, function() {});

    var processMoveLine = function(line) {
	//sleepFor(100)
	console.log(line)
	try {
	    let l = line.split(' ')
	    let p = l[0]
	    let s = l[1]
	    let t = l[2]
	    updateEdgeOwner(p, s, t)
	} catch (u) {
	    console.log("?!" + u)
	}
    }

    for (var i = 1; i < lines.length; i++) {
	//processMoveLine(lines[i])
	//setTimeout(function() { processMoveLine(lines[i]) }.bind(lines), 100)
	setTimeout(processMoveLine, 10*i, lines[i])
    }
}

function visualizeFromFile(logFile) {
    var reader = new FileReader();
    reader.onload = function(progressEvent) {
	var lines = this.result.split('\n');
	visualize(lines)
    };
    reader.readAsText(logFile);
}

document.getElementById('visfile').onchange = function() {
    visualizeFromFile(this.files[0])
};
