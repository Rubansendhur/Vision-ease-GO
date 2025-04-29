// Quiz functionality
document.addEventListener('DOMContentLoaded', function() {
    // Quiz functionality
    const quizLevel = document.getElementById('quizLevel');
    const startQuizButton = document.getElementById('startQuiz');
    const quizContainer = document.getElementById('quizContainer');
    const quizContent = document.getElementById('quizContent');
    const quizResults = document.getElementById('quizResults');
    let currentQuizzes = [];

    startQuizButton.addEventListener('click', function() {
        const level = quizLevel.value;
        loadQuizzes(level);
    });

    function loadQuizzes(level) {
        fetch(`/api/quizzes?level=${level}`)
            .then(response => response.json())
            .then(quizzes => {
                if (quizzes.message) {
                    // No quizzes found
                    quizContent.innerHTML = `<p class="error">${quizzes.message}</p>`;
                    return;
                }

                currentQuizzes = quizzes;
                displayQuizzes(quizzes);
                quizContainer.classList.remove('hidden');
            })
            .catch(error => {
                quizContent.innerHTML = `<p class="error">Error loading quizzes: ${error.message}</p>`;
            });
    }

    function displayQuizzes(quizzes) {
        const quizHTML = quizzes.map((quiz, index) => `
            <div class="question" data-quiz-id="${quiz.ID}">
                <h3>Question ${index + 1}</h3>
                <p>${quiz.Question}</p>
                <div class="options">
                    ${quiz.Options.map(option => `
                        <label>
                            <input type="radio" name="q${index}" value="${option}" required>
                            ${option}
                        </label>
                    `).join('')}
                </div>
            </div>
        `).join('');

        quizContent.innerHTML = `
            <form id="quizForm">
                ${quizHTML}
                <button type="submit" class="submit-button">Submit Answers</button>
            </form>
        `;

        // Add form submit handler
        document.getElementById('quizForm').addEventListener('submit', function(event) {
            event.preventDefault();
            checkAnswers(quizzes);
        });
    }

    function checkAnswers(quizzes) {
        let score = 0;
        const results = [];

        quizzes.forEach((quiz, index) => {
            const selectedAnswer = document.querySelector(`input[name="q${index}"]:checked`)?.value;
            const isCorrect = selectedAnswer === quiz.Answer;
            if (isCorrect) score++;

            results.push({
                question: quiz.Question,
                selectedAnswer: selectedAnswer,
                correctAnswer: quiz.Answer,
                explanation: quiz.Explanation,
                isCorrect: isCorrect
            });
        });

        displayResults(score, quizzes.length, results);
    }

    function displayResults(score, total, results) {
        const percentage = (score / total) * 100;
        let message = '';
        if (percentage === 100) {
            message = 'Perfect! You know your stuff!';
        } else if (percentage >= 70) {
            message = 'Great job! You have a good understanding of color blindness.';
        } else if (percentage >= 40) {
            message = 'Not bad! Keep learning about color blindness.';
        } else {
            message = 'Keep studying! There\'s more to learn about color blindness.';
        }

        const resultsHTML = results.map(result => `
            <div class="result-item ${result.isCorrect ? 'correct' : 'incorrect'}">
                <h4>${result.question}</h4>
                <p>Your answer: ${result.selectedAnswer || 'Not answered'}</p>
                <p>Correct answer: ${result.correctAnswer}</p>
                <p class="explanation">${result.explanation}</p>
            </div>
        `).join('');

        quizResults.innerHTML = `
            <h3>Quiz Results</h3>
            <p>You scored ${score} out of ${total} (${percentage.toFixed(1)}%)</p>
            <p>${message}</p>
            <div class="detailed-results">
                ${resultsHTML}
            </div>
            <button class="retry-button" onclick="location.reload()">Try Another Level</button>
        `;

        quizResults.classList.remove('hidden');
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
