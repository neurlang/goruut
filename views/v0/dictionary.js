// Global array to store word objects
let wordsArray = [];

// Method to push a new word object into the array
function addWord(cleanWord, prePunct, phonetic, postPunct) {
    wordsArray.push({
        CleanWord: cleanWord,
        PrePunct: prePunct,
        Phonetic: phonetic,
        PostPunct: postPunct
    });
}

// Method to repaint the DataTable
function repaintDataTable(dataTable) {
    // Clear the existing DataTable
    dataTable.destroy();

    // Reinitialize the DataTable with updated data
    const table = document.getElementById('wordsTable');
    dataTable = new simpleDatatables.DataTable(table, {
        data: {
            headings: ["Select", "Clean Word", "Pre Punctuation", "Phonetic", "Post Punctuation"],
            data: wordsArray.map(word => [
                `<input type="checkbox" class="word-checkbox" data-word="${word.CleanWord}" data-phonetic="${word.Phonetic}">`,
                word.CleanWord,
                word.PrePunct,
                word.Phonetic,
                word.PostPunct
            ])
        },
        columns: [
            { select: 0, sortable: false }, // Checkbox column (not sortable)
            { select: 1 }, // Clean Word
            { select: 2 }, // Pre Punctuation
            { select: 3 }, // Phonetic
            { select: 4 }  // Post Punctuation
        ]
    });

    // Store the DataTable instance for later use
    window.dataTable = dataTable;
}

// Initialize DataTable on page load
document.addEventListener('DOMContentLoaded', function () {
    // Initialize the DataTable
    const table = document.getElementById('wordsTable');
    const dataTable = new simpleDatatables.DataTable(table, {
        data: {
            headings: ["Select", "Clean Word", "Pre Punctuation", "Phonetic", "Post Punctuation"],
            data: [] // Initially empty
        },
        columns: [
            { select: 0, sortable: false }, // Checkbox column (not sortable)
            { select: 1 }, // Clean Word
            { select: 2 }, // Pre Punctuation
            { select: 3 }, // Phonetic
            { select: 4 }  // Post Punctuation
        ]
    });

    // Store the DataTable instance for later use
    window.dataTable = dataTable;

    // Event listener for delete button
    document.getElementById('deleteSelected').addEventListener('click', function () {
        let deduplicatedArray = wordsArray;
    
        // Find all checked checkboxes
        document.querySelectorAll('.word-checkbox:checked').forEach(checkbox => {
            const wordToDelete = checkbox.dataset.word;
            const pwordToDelete = checkbox.dataset.phonetic;

            // Remove the word from the global array
            deduplicatedArray = deduplicatedArray.filter(word => word.CleanWord != wordToDelete || word.Phonetic != pwordToDelete);
            
        });
                
        wordsArray = deduplicatedArray;

        // Repaint the DataTable after deletion
        repaintDataTable(window.dataTable);
    });

    // Example usage:
    // Add some words to the array
    //addWord("Hello", "", "həˈloʊ", "!");
    //addWord("World", "", "wɜːrld", ".");

    // Repaint the DataTable to reflect the new words
    //repaintDataTable(dataTable);
});

/**
 * Deduplicates the `wordsArray` based on all four attributes.
 * Two records are considered equal if all four strings are equal.
 */
function deduplicateWordsArray() {
    // Create a Set to store unique string representations of the objects
    const uniqueSet = new Set();

    // Filter the array to keep only unique objects
    const deduplicatedArray = wordsArray.filter(word => {
        // Create a unique string key for the object
        const key = `${word.CleanWord}|${word.PrePunct}|${word.Phonetic}|${word.PostPunct}`;

        // If the key is not in the Set, add it and keep the object
        if (!uniqueSet.has(key)) {
            uniqueSet.add(key);
            return true;
        }

        // Otherwise, discard the object (it's a duplicate)
        return false;
    });

    wordsArray = deduplicatedArray;
}

/**
 * Searches the DataTable for a specific string using the library's built-in search
 * @param {string} searchTerm - The term to search for
 */
function searchTable(searchTerm) {
    if (window.dataTable) {
        window.dataTable.search(searchTerm);
        // Get the filtered rows
	const list = wordsArray.filter(word => word.Phonetic === searchTerm).map(word => word.CleanWord); // Extract only the `CleanWord` values
	return list.length === 0 ? '' : list[0];
    }
    return '';
}
