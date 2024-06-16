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

document.getElementById('phonemizer').onclick = function() {
    // Extract text from the textarea with id "text"
    const text = document.getElementById('textt').value;
    const lang = document.getElementById('language').value;

    // Define the data to be sent in the POST request
    const data = {
        "Language": lang,
        "IpaFlavors": [],
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
    	document.getElementById('output').innerText = '';
    	for (var i in data.Words) {
    	const word = data.Words[i];
    	const lang = document.getElementById('output').innerText += " " + word.Phonetic;
    	}
        console.log('Success:', data);
    })
    .catch((error) => {
        console.error('Error:', error);
    });
};
