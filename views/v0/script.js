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

    var target = [];
    if (targ == "Espeak") {
	target = ["Espeak", "Espeak_"+lang];
    } else if (targ == "Antvaset") {
	target = ["antvaset.com", "antvaset.com_"+lang];
    }

    // Define the data to be sent in the POST request
    const data = {
        "Language": lang,
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
    	document.getElementById('output').value = '';
    	for (var i in data.Words) {
	const word = data.Words[i];
	const lang = document.getElementById('output').value += ((targ == "Antvaset") ? "" : " ") + word.Phonetic;
    	}
        console.log('Success:', data);
    })
    .catch((error) => {
        console.error('Error:', error);
    });
};
