
function toggleSearch() {
    const searchArea = document.getElementById('searchArea');
    const searchIcon = document.getElementById('search-icon');
    if (searchArea.classList.contains('show')) {
        searchArea.classList.remove('show');
        searchIcon.style.display = 'inline';
    } else {
        searchArea.classList.add('show');
        searchIcon.style.display = 'none';
    }
  }
  
  // Filter function to show/hide elements based on the input value
  function filterContent(event) {
    const filterText = event.target.value.toLowerCase(); // Get search input text in lowercase
    const contentElements = document.querySelectorAll('.filtrable'); // Get all elements with class 'filtrable'

  
    contentElements.forEach(element => {
        const text = element.textContent.toLowerCase();
        if (text.includes(filterText)) {
            element.style.display = '';
        } else {
            element.style.display = 'none';
        }
    });
  }

  document.getElementById('searchArea').addEventListener('input', filterContent);
  