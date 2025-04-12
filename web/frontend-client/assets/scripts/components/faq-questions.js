const faqQuestions = document.querySelector('[data-element="faq-questions"]')

if (faqQuestions) setTimeout(faqInit, 0)

function faqInit () {
  const questions = faqQuestions.querySelectorAll('[data-element="faq__question"]')

  if (questions.length) {
    for (let i = 0; i < questions.length; i++) {
      questions[i].addEventListener('click', showContent)
    }
  }

  function showContent () {
    const parent = this.parentNode

    if (parent) {
      parent.classList.toggle('faq-questions__item-active')
    }
  }
}
