export { animateScrollToAnchor }
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
