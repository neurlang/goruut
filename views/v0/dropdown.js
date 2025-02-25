function initDropdown(inp, lst, items) {

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
          searchInput.value = item;
          dropdownList.classList.remove("show");
        });
        dropdownList.appendChild(itemDiv);
      });
    }
  });

  // Event listener for input changes
  searchInput.addEventListener("input", function () {
    const filter = searchInput.value.toLowerCase();
    dropdownList.innerHTML = ""; // Clear the list

    // Filter items based on input and update dropdown
    const filteredItems = items.filter(item => item.toLowerCase().includes(filter));

    if (filteredItems.length > 0) {
      dropdownList.classList.add("show"); // Show dropdown if there are items
      filteredItems.forEach(item => {
        const itemDiv = document.createElement("div");
        itemDiv.textContent = item;
        itemDiv.addEventListener("click", () => {
          searchInput.value = item;
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
"English", "Esperanto", "Estonian", "Farsi", "Finnish", "French", "Galician", "Georgian", "German", "Greek", "Gujarati", "Hausa",
"Hebrew", "Hindi", "Hungarian", "Icelandic", "Indonesian", "Italian", "Jamaican", "Japanese",
"Javanese", "Kazakh", "KhmerCentral", "Korean", "Lao", "Latvian", "Lithuanian", "Luxembourgish", "Macedonian", "Malayalam", "MalayLatin",
"Maltese", "Marathi", "Mongolian", "Nepali", "Norwegian", "Pashto", "Polish", "Portuguese",
"Punjabi", "Romanian", "Russian", "Serbian", "Slovak", "Spanish", "Swahili", "Swedish", "Tagalog", "Tamil", "Telugu",
"Thai", "Tibetan", "Turkish", "Ukrainian", "Urdu", "Uyghur", "VietnameseNorthern", "Yoruba", "Zulu",
"Isan", "BengaliDhaka", "BengaliRahr", "MalayArab", "VietnameseCentral", "VietnameseSouthern"];
initDropdown("langsearchInput", "langdropdownList", items);
const reverse_items = ["IPA", "Espeak", "Antvaset"].concat(items);
initDropdown("tgtsearchInput", "tgtdropdownList", reverse_items);
