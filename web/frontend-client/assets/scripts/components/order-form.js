const feedbackForm = document.querySelector('[data-role="feedback-form"]');

if (feedbackForm) setTimeout(initFeedbackForm, 0);

function initFeedbackForm() {
  const forms = document.querySelectorAll('[data-role="feedback-form"]');

  forms.forEach((form) => {
    const phoneInput = form.querySelector(
      '[data-role="feedback-form__phone-input"]'
    );
    const errorText = form.querySelector(
      '[data-role="feedback-form__text-error"]'
    );
    const submitButton = form.querySelector(
      '[data-role="feedback-form__button"]'
    );

    if (!phoneInput || !errorText || !submitButton) return;

    const phoneRegex = new RegExp(phoneInput.pattern);

    phoneInput.addEventListener("input", validatePhone);
    phoneInput.addEventListener("blur", validatePhone);

    form.addEventListener("submit", (e) => {
      if (!validatePhone()) {
        e.preventDefault();
      }
    });

    function validatePhone() {
      const phoneValue = phoneInput.value.trim();
      const isValid = phoneValue.length >= 6 && phoneRegex.test(phoneValue);

      if (!isValid && phoneValue.length > 0) {
        phoneInput.classList.add("input_error");
        errorText.classList.remove("text-error_hide");
        submitButton.disabled = true;
        return false;
      } else {
        phoneInput.classList.remove("input_error");
        errorText.classList.add("text-error_hide");
        submitButton.disabled = false;
        return true;
      }
    }
  });
}
