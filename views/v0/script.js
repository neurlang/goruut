function unicodeEscapeNonLatin(text) {
    return text.split('').map(char => {
        // Check if the character code is outside the Latin range
        if (char.charCodeAt(0) > 127) {
            // Convert the character to its Unicode escape sequence
            return '\\u' + char.charCodeAt(0).toString(16).padStart(4, '0');
        }
        return char;
    }).join('');
}

document.getElementById('copyit').onclick = function() {
	const text = document.getElementById('output').value;
	navigator.clipboard.writeText(text);
}

document.getElementById('phonemizer').onclick = function() {
    // Extract text from the textarea with id "text"
    const text = document.getElementById('textt').value;
    const langid = document.getElementById('langsearchInput');
    const targid = document.getElementById('tgtsearchInput');
    
    const lang = langid === null ? "" : langid.value;
    const targ = targid === null ? "" :  targid.value;

    var singleLookup = false;
    var target = [];
    if (targ == "Espeak") {
	target = ["Espeak", "Espeak_"+lang];
	singleLookup = true;
    } else if (targ == "Antvaset") {
	target = ["antvaset.com", "antvaset.com_"+lang];
	singleLookup = true;
    } else if (targ == "IPA") {
        singleLookup = true;
    }

    // Define the data to be sent in the POST request
    const data = {
        "Language": lang.includes(',') ? "" : lang,
        "Languages": lang.includes(',') ? lang.replace(/^,+|,+$/g, '').split(',') : [],
        "IpaFlavors": target,
        "Sentence": text
    };

    // Send the POST request to the specified endpoint
    fetch('/tts/phonemize/sentence', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: unicodeEscapeNonLatin(JSON.stringify(data))
    })
    .then(response => response.json())
    .then(data => {
    	if (singleLookup) {
	    	document.getElementById('output').innerHTML = '';
	    	for (var i in data.Words) {
		const word = data.Words[i];
		const not_dict = !word.PosTags?.includes("dict");
		const ubegin = not_dict ? "<u>" : "";
		const uend = not_dict ? "</u>" : "";
		const lang = document.getElementById('output').innerHTML +=
			((targ == "Antvaset") ? "" : " ") + "<b>" + word.PrePunct + "</b>" + ubegin + word.Phonetic + uend + "<b>" + word.PostPunct + "</b>";
	    	}
		console.log('Success:', data);
		const checkbox = document.getElementById('toggleCheckbox');
		checkbox.click();
		checkbox.click();
		return
        }
            partial = "";
	    for (var i in data.Words) {
		const word = data.Words[i];
		partial += " " + word.PrePunct + word.Phonetic + word.PostPunct;
	    }
	    // Define the data to be sent in the POST request
	    const data2 = {
	    	"IsReverse": true,
		"Language": targ,
		"IpaFlavors": [],
		"Sentence": partial
	    };
	    fetch('/tts/phonemize/sentence', {
		method: 'POST',
		headers: {
		    'Content-Type': 'application/json'
		},
		body: unicodeEscapeNonLatin(JSON.stringify(data2))
	    })
	    .then(response => response.json())
	    .then(data => {
	    	document.getElementById('output').innerHTML = '';
	    	for (var i in data.Words) {
		const word = data.Words[i];
		const not_dict = !word.PosTags?.includes("dict");
		const ubegin = not_dict ? "<u>" : "";
		const uend = not_dict ? "</u>" : "";
		const lang = document.getElementById('output').innerHTML +=
			((targ == "Antvaset") ? "" : " ") + "<b>" + word.PrePunct + "</b>" + ubegin + word.Phonetic + uend + "<b>" + word.PostPunct + "</b>";
	    	}
		console.log('Success:', data);
		const checkbox = document.getElementById('toggleCheckbox');
		checkbox.click();
		checkbox.click();
		return
	    })
	    .catch((error) => {
		console.error('Error:', error);
	    });
    })
    .catch((error) => {
        console.error('Error:', error);
    });
};
