// Quiz functionality
document.addEventListener('DOMContentLoaded', function() {
    const quizForm = document.getElementById('quizForm');
    if (quizForm) {
        quizForm.addEventListener('submit', function(event) {
            event.preventDefault();
            
            const answers = {
                q1: document.querySelector('input[name="q1"]:checked')?.value,
                q2: document.querySelector('input[name="q2"]:checked')?.value,
                q3: document.querySelector('input[name="q3"]:checked')?.value
            };

            const correctAnswers = {
                q1: 'protanopia',
                q2: 'male',
                q3: 'correction'
            };

            let score = 0;
            let feedback = '';

            for (const [question, answer] of Object.entries(answers)) {
                if (answer === correctAnswers[question]) {
                    score++;
                }
            }

            const resultsDiv = document.getElementById('quizResults');
            resultsDiv.innerHTML = `
                <h3>Quiz Results</h3>
                <p>You scored ${score} out of 3!</p>
                <p>${score === 3 ? 'Perfect! You know your stuff!' : 'Keep learning about color blindness!'}</p>
            `;
        });
    }

    // Image upload and processing functionality
    const uploadForm = document.getElementById('uploadForm');
    if (uploadForm) {
        const imageInput = document.getElementById('imageInput');
        const imagePreview = document.getElementById('imagePreview');
        const previewImage = document.getElementById('previewImage');
        const loadingIndicator = document.getElementById('loadingIndicator');
        const uploadButton = document.querySelector('.upload-button');
        const resultsSection = document.getElementById('results');

        // Show/hide rotation options based on transformation selection
        const transformationSelect = document.getElementById('transformation');
        const rotationOptions = document.getElementById('rotationOptions');
        const angleInput = document.getElementById('angle');
        const angleValue = document.getElementById('angleValue');

        // Handle image preview
        imageInput.addEventListener('change', function() {
            const file = this.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    previewImage.src = e.target.result;
                    imagePreview.classList.remove('hidden');
                    // Clear previous results when a new image is selected
                    resultsSection.innerHTML = '';
                };
                reader.readAsDataURL(file);
            } else {
                imagePreview.classList.add('hidden');
                resultsSection.innerHTML = '';
            }
        });

        transformationSelect.addEventListener('change', function() {
            if (this.value === 'rotate' || this.value === 'rotate_shear') {
                rotationOptions.classList.remove('hidden');
            } else {
                rotationOptions.classList.add('hidden');
            }
        });

        angleInput.addEventListener('input', function() {
            angleValue.textContent = this.value + '°';
        });

        uploadForm.addEventListener('submit', function(event) {
            event.preventDefault();
            
            // Show loading indicator and disable upload button
            loadingIndicator.classList.remove('hidden');
            uploadButton.disabled = true;
            
            const formData = new FormData(uploadForm);
            
            // Get selected options
            const colorBlindness = document.getElementById('colorBlindness').value;
            const transformation = document.getElementById('transformation').value;
            const filter = document.getElementById('filter').value;
            const angle = document.getElementById('angle').value;

            // Build the URL with query parameters
            let url = '/visualize?';
            const operations = [];

            // Add selected operations
            if (colorBlindness !== 'none') {
                operations.push(colorBlindness);
            }
            if (transformation !== 'none') {
                operations.push(transformation);
            }
            if (filter !== 'none') {
                operations.push(filter);
            }

            // Add all operations to the URL
            operations.forEach(op => {
                url += `operation=${op}&`;
            });

            // Add angle if needed
            if (transformation === 'rotate' || transformation === 'rotate_shear') {
                url += `angle=${angle}&`;
            }

            fetch(url, {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                resultsSection.innerHTML = `
                    <h3>Processing Results</h3>
                    <div class="results-grid">
                        ${data.images.map((url, index) => `
                            <div class="result-card">
                                <h4>${index === 0 ? 'Original' : `Step ${index}: ${data.operations[index-1]}`}</h4>
                                <img src="${url}" alt="Processed image">
                            </div>
                        `).join('')}
                    </div>
                    <div class="applied-operations">
                        <h4>Applied Operations:</h4>
                        <ul>
                            ${operations.map(op => `<li>${op}</li>`).join('')}
                        </ul>
                    </div>
                `;
            })
            .catch(error => {
                resultsSection.innerHTML = `
                    <p class="error">Error processing image: ${error.message}</p>
                `;
            })
            .finally(() => {
                // Hide loading indicator and enable upload button
                loadingIndicator.classList.add('hidden');
                uploadButton.disabled = false;
            });
        });
    }
});
