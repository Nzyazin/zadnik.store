export function exportHeader() {
  if (document.querySelector('[data-element="header-mob__burger"]')) {
    setTimeout(headerInit, 0);
  }
}

function headerInit () {
  let isMenuOpen = false;
  let layer = document.querySelector('[data-element="header__layer"]');
  let buttonBurger = document.querySelector('[data-element="header-mob__burger"]');
  let menu = document.querySelector('[data-element="menu"]');
  buttonBurger.addEventListener('click', toogleMenu);
  layer.addEventListener('click', hideMenu);

  function hideMenu () {
    closeMenu();
    hideLayer();
    isMenuOpen = false
  }

  function showMenu () {
    console.log('dssds')
    openMenu();
    showLayer();
    isMenuOpen = true
  }

  function toogleMenu() {
    if (isMenuOpen) {
      hideMenu()
    } else {
      showMenu()
    }
  }

  function openMenu () {menu.classList.add('header__menu-mobile_active')}
  function closeMenu () {menu.classList.remove('header__menu-mobile_active')}
  function showLayer () {layer.classList.add('header__layer_show')}
  function hideLayer () {layer.classList.remove('header__layer_show')}
}
