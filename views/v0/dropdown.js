const checkbox = document.getElementById('toggleCheckbox');
// Add an event listener to the checkbox
checkbox.addEventListener('change', function() {
  // Get references to the checkbox and the div with the ID 'output'
  const outputDiv = document.getElementById('output');
  // Get the <b> tag inside the output div
  const boldTags = outputDiv.querySelectorAll('b');
  
  // Loop through all <b> tags and toggle their visibility
  boldTags.forEach(boldTag => {
    if (this.checked) {
      boldTag.style.display = 'inline';
    } else {
      boldTag.style.display = 'none';
    }
  });
});

function initDropdown(inp, lst, items, multiple) {

  // Get the dropdown elements
  const searchInput = document.getElementById(inp);
  const dropdownList = document.getElementById(lst);
  // Event listener for input changes
  searchInput.addEventListener("click", function () {
    dropdownList.innerHTML = ""; // Clear the list
    
    if (items.length > 0) {
      dropdownList.classList.add("show"); // Show dropdown if there are items
      items.forEach(item => {
        const itemDiv = document.createElement("div");
        itemDiv.textContent = item;
        itemDiv.addEventListener("click", () => {
          if (multiple) {
            searchInput.value += item + ",";
          } else {
            searchInput.value = item;
          }
          dropdownList.classList.remove("show");
        });
        dropdownList.appendChild(itemDiv);
      });
    }
  });

  // Event listener for input changes
  searchInput.addEventListener("input", function () {
    var filter = searchInput.value.toLowerCase();
    if (multiple && filter.includes(',')) {
      const spl = filter.split(',');
      filter = spl[spl.length-1];
    }
    dropdownList.innerHTML = ""; // Clear the list

    // Filter items based on input and update dropdown
    const filteredItems = items.filter(item => item.toLowerCase().includes(filter));

    if (filteredItems.length > 0) {
      dropdownList.classList.add("show"); // Show dropdown if there are items
      filteredItems.forEach(item => {
        const itemDiv = document.createElement("div");
        itemDiv.textContent = item;
        itemDiv.addEventListener("click", () => {
          if (multiple) {
		if (searchInput.value.endsWith(filter)) {
		    searchInput.value = searchInput.value.slice(0, -filter.length);
		}
            searchInput.value += item + ",";
          } else {
            searchInput.value = item;
          }
          dropdownList.classList.remove("show");
        });
        dropdownList.appendChild(itemDiv);
      });
    } else {
      dropdownList.classList.remove("show"); // Hide if no items match
    }
  });

  // Hide dropdown when clicking outside
  document.addEventListener("click", (event) => {
    if (!searchInput.contains(event.target) && !dropdownList.contains(event.target)) {
      dropdownList.classList.remove("show");
    }
  });

}

// Sample data for dropdown items
const items = ["Afrikaans", "Amharic", "Arabic", "Armenian", "Azerbaijani", "Basque", "Belarusian", "Bengali", "Bulgarian",
"Burmese", "Cebuano", "Chechen", "Chichewa", "ChineseMandarin", "Catalan", "Croatian", "Czech", "Danish", "Dutch", "Dzongkha",
"English", "EnglishAmerican", "EnglishBritish", "Esperanto", "Estonian", "Farsi", "Finnish", "French", "Galician", "Georgian",
"German", "Greek", "Gujarati", "Hausa", "Hebrew", "Hindi", "Hungarian", "Icelandic", "Indonesian", "Italian", "Jamaican", "Japanese",
"Javanese", "Kazakh", "KhmerCentral", "Korean", "Lao", "Latvian", "Lithuanian", "Luxembourgish", "Macedonian", "Malayalam", "MalayLatin",
"Maltese", "Marathi", "Mongolian", "Nepali", "Norwegian", "Pashto", "Polish", "Portuguese",
"Punjabi", "Romanian", "Russian", "Serbian", "Slovak", "Spanish", "Swahili", "Swedish", "Tagalog", "Tamil", "Telugu",
"Thai", "Tibetan", "Turkish", "Ukrainian", "Urdu", "Uyghur", "VietnameseNorthern", "Yoruba", "Zulu",
"Isan", "BengaliDhaka", "BengaliRahr", "MalayArab", "VietnameseCentral", "VietnameseSouthern"];
initDropdown("langsearchInput", "langdropdownList", items, true);
const reverse_items = ["IPA", "Espeak", "Antvaset"].concat(items);
initDropdown("tgtsearchInput", "tgtdropdownList", reverse_items, false);
