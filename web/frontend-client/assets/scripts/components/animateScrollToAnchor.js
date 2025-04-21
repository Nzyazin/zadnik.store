if (document.querySelector('[data-role="scroll-to-anchor"]')) {
  setTimeout(() => {
    const anchorElements = document.querySelectorAll(
      '[data-role="scroll-to-anchor"]'
    );

    for (let i = 0, len = anchorElements.length; i < len; i++)
      _loopAddEventScrollToAnchor(i);

    function _loopAddEventScrollToAnchor(theIndexNode) {
      anchorElements[theIndexNode].addEventListener(
        "click",
        clickOnTheScrollElement
      );
    }

    function clickOnTheScrollElement(event) {
      event.preventDefault();
      let elementId;
      if (this.dataset.link) elementId = this.dataset.link.substr(1);
      else elementId = this.hash.substr(1);
      const element = document.getElementById(elementId);
      if (element) animateScrollToAnchor(element);
    }
  }, 0);


function animateScrollToAnchor(theElement) {
  const positionNow = window.pageYOffset;
  const positionElement =
    theElement.getBoundingClientRect().top + pageYOffset - 50;
  const duration = 300;
  const step = positionElement - positionNow;
  const start = performance.now();

  requestAnimationFrame(function animate(time) {
    const timePassed = time - start;

    if (timePassed > duration) {
      window.scrollTo(0, positionElement);
    } else {
      window.scrollTo(0, positionNow + step * (timePassed / duration));
      requestAnimationFrame(animate);
    }
  });
}
}
