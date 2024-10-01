
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
  
  
  function filterContent(event) {
    const filterText = event.target.value.toLowerCase(); // Get search input text in lowercase
    const contentElements = document.querySelectorAll('.filtrable');
  
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
  