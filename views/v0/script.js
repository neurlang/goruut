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
    const lng = lang.replace(/^,+|,+$/g, '');
    const data = {
        "Language": lng.includes(',') ? "" : lng,
        "Languages": lng.includes(',') ? lng.split(',') : [],
        "IpaFlavors": target,
        "Sentence": text,
        "SplitSentences": true
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
		const is_last_hr = word.IsLast ? "<hr />" : "";
		const ubegin = not_dict ? "<u>" : "";
		const uend = not_dict ? "</u>" : "";
		const lang = document.getElementById('output').innerHTML +=
			((targ == "Antvaset") ? "" : " ") + "<b>" + word.PrePunct + "</b>" + ubegin + word.Phonetic + uend + "<b>" + word.PostPunct + "</b>" + is_last_hr;
			if (not_dict) { addWord(word.CleanWord, word.PrePunct, word.Phonetic, word.PostPunct); }
	    	}
		console.log('Success:', data);
		const checkbox = document.getElementById('toggleCheckbox');
		checkbox.click();
		checkbox.click();
		const checkbox2 = document.getElementById('toggleSentCheckbox');
		checkbox2.click();
		checkbox2.click();
		deduplicateWordsArray();
		repaintDataTable(window.dataTable);
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
		const is_last_hr = word.IsLast ? "<hr />" : "";
		const ubegin = not_dict ? "<u>" : "";
		const uend = not_dict ? "</u>" : "";
		const lang = document.getElementById('output').innerHTML +=
			((targ == "Antvaset") ? "" : " ") + "<b>" + word.PrePunct + "</b>" + ubegin + word.Phonetic + uend + "<b>" + word.PostPunct + "</b>" + is_last_hr;
			if (not_dict) { addWord(word.CleanWord, word.PrePunct, word.Phonetic, word.PostPunct); }
	    	}
		console.log('Success:', data);
		const checkbox = document.getElementById('toggleCheckbox');
		checkbox.click();
		checkbox.click();
		const checkbox2 = document.getElementById('toggleSentCheckbox');
		checkbox2.click();
		checkbox2.click();
		deduplicateWordsArray();
		repaintDataTable(window.dataTable);
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

function findAndAddLongestPrefixToTable(word, map, otherword) {
    const table = document.getElementById('wordxplain');
    
    content = [];
    extras = [];
    original = [];

    // Helper function to process word and otherword
    function process(word, otherword) {
        if ((word.length == 0) && (otherword.length == 0)) {return true;}
        if ((word.length == 0)) {return false;}
        if (otherword.length == 0) {return false;}
        
        {
            let longestPrefix = '';
            let longestValue = [];

            // Find the longest prefix in the map for the current word
            for (const prefix in map) {
                if (word.startsWith(prefix) && prefix.length > longestPrefix.length) {
                    longestPrefix = prefix;
                    longestValue = map[prefix];
                }
            }

            // If no prefix is found, break the loop
            if (longestPrefix === '') {
                return false;
            }
            let buffer = [];
            for (let i in longestValue) {
            buffer.push(longestValue[i]);
            // Check if the value is a prefix of otherword
            if (otherword.startsWith(longestValue[i])) {
                
                // Consume the longest prefix from word and the matching prefix from otherword
                if (process(word.slice(longestPrefix.length), otherword.slice(longestValue[i].length))) {

                    content.push(buffer);
                    original.push(longestPrefix);
                    extras.push(longestValue.slice(i+1));
                    return true;
                }
            }
            }
            return process(word.slice(longestPrefix.length), otherword);
        }
    }

    // Start the process
    const success = process(word, otherword);
    if (!success) {
    	return false;
    }
    table.innerHTML = ''; // Clear the table before starting
    
    let maximum = 0;
    for (const i in content) {
        if (content[i].length > maximum) {
            maximum = content[i].length;
        }
    }
    let past = '';
    let future = '';
    for (const i in content) {
	future += content[content.length-i-1][content[content.length-i-1].length-1];
    }
    for (const i in content) {
        future = future.slice(content[content.length-i-1][content[content.length-i-1].length-1].length);
	const row = table.insertRow();
	const headercell = row.insertCell();
	headercell.textContent = original[content.length-i-1];
	headercell.classList.add('w3-blue');
        for (let j = 0; j < i; j++) {const cell = row.insertCell(); cell.textContent = '';}
	for (let j = maximum; j > content[content.length-i-1].length;j--) {const cell = row.insertCell(); cell.textContent = '';}
    	for (let j in content[content.length-i-1]) {
	    const cell = row.insertCell();
	    cell.textContent = content[content.length-i-1][j];
	    if ((j==content[content.length-i-1].length-1) || (content[content.length-i-1].length == 1)) {
	        // Highlight the cell if it forms the alignment
	        cell.classList.add('w3-green');
	    } else {
	        cell.classList.add('w3-red');
	        const str = past + content[content.length-i-1][j] + future;
	        cell.onclick = () => toggleWordAndIpa(word, str);
	    }
	}
	for (let j in extras[extras.length-i-1]) {
	    const cell = row.insertCell();
	    cell.textContent = extras[extras.length-i-1][j];
	    cell.classList.add('w3-deep-orange');
	    const str = past + extras[extras.length-i-1][j] + future;
	    cell.onclick = () => toggleWordAndIpa(word, str);
	}
	past += content[content.length-i-1][content[content.length-i-1].length-1];
    }
    return true;
}


let wordpair_rules = null;

function loadWordAndIpa(txt, ipa) {
	const langid = document.getElementById('langsearchInput');
	const lang = langid === null ? "" : langid.value;
	const lng = lang.replace(/^,+|,+$/g, '');
	const data2 = {
	    	"CleanWord": txt,
		"Phonetic": ipa,
		"IsReverse": false,
		"Language": lng.includes(',') ? lng.split(',')[0] : lng
	};
	fetch('/tts/explain/word', {
		method: 'POST',
		headers: {
		    'Content-Type': 'application/json'
		},
		body: unicodeEscapeNonLatin(JSON.stringify(data2))
	})
	.then(response => response.json())
	.then(data => {
		wordpair_rules = data['Rules'];
		findAndAddLongestPrefixToTable(txt, data['Rules'], ipa);
	})
	.catch((error) => {
	console.error('Error:', error);
	});
}

document.getElementById('output').onclick = function(x, y) {
	const clickedElement = event.target;
	const ipa = clickedElement.innerText;
	if (ipa == '') {return;}
	const txt = searchTable(ipa);
	if (txt == '') {return;}
	const wordpair = document.getElementById('wordpair');
	wordpair.innerText = txt + " [" + ipa + "]";
	
	loadWordAndIpa(txt, ipa);
}

function toggleWordAndIpa(word, ipa) {
	const wordpair = document.getElementById('wordpair');
	wordpair.innerText = word + " [" + ipa + "]";
	if (!findAndAddLongestPrefixToTable(word, wordpair_rules, ipa)) {
		loadWordAndIpa(word, ipa);
	}
}
