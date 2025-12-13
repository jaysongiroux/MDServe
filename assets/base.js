/*
Base JavaScript for MDServe
*/

// Add copy buttons to code blocks
document.addEventListener('DOMContentLoaded', function() {
  // Find all code elements in page content
  const codeBlocks = document.querySelectorAll('.page-content code');
  
  codeBlocks.forEach(function(code) {
    // Check if it's multi-line by counting line breaks
    const text = code.textContent || code.innerText;
    const lineCount = text.split('\n').length;
    
    // Only add button for multi-line code blocks (more than 1 line)
    if (lineCount <= 1) return;
    
    // Skip if already wrapped
    if (code.parentElement && code.parentElement.classList.contains('code-block-wrapper')) {
      return;
    }
    
    // Create wrapper
    const wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';
    code.parentNode.insertBefore(wrapper, code);
    wrapper.appendChild(code);
    
    // Create copy button
    const button = document.createElement('button');
    button.className = 'copy-code-button';
    button.setAttribute('aria-label', 'Copy code to clipboard');
    button.setAttribute('title', 'Copy code');
    button.innerHTML = `
      <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
      </svg>
      <svg class="check-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="20 6 9 17 4 12"></polyline>
      </svg>
    `;
    
    // Add click handler
    button.addEventListener('click', function() {
      const codeText = getCodeText(code);
      
      navigator.clipboard.writeText(codeText).then(function() {
        // Show success state
        button.classList.add('copied');
        
        // Reset after 2 seconds
        setTimeout(function() {
          button.classList.remove('copied');
        }, 2000);
      }).catch(function(err) {
        console.error('Failed to copy code:', err);
      });
    });
    
    wrapper.insertBefore(button, code);
  });
  
  // Helper function to extract code text
  function getCodeText(codeElement) {
    // Simply return the text content - syntax highlighting spans don't affect textContent
    return codeElement.textContent || codeElement.innerText;
  }
});