if (document.querySelector('[data-element="cookies"]')) {
  setTimeout(initCookies, 0)
}

function initCookies () {
  const cookiesName = 'NDrssC'
  const asideCookies = document.querySelector('[data-element="cookies"]')
  const buttonClose = document.querySelector('[data-element="cookies__button"]')

  function checkCookies () {
    const cookie = getCookies()
    if (!cookie) showCookiesPanel()
  }

  checkCookies()

  function showCookiesPanel () {
    asideCookies.classList.add('cookies_show')
  }

  function getCookies () {
    return localStorage.getItem(cookiesName)
  }

  function setCookies () {
    localStorage.setItem(cookiesName, '1')
  }

  function permissionCookies () {
    setCookies()
    hideCookiesPanel()
  }

  function hideCookiesPanel () {
    asideCookies.classList.remove('cookies_show')
  }

  buttonClose.addEventListener("click", permissionCookies)
}
