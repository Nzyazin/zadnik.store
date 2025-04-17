import { exportHeader } from "./components/header.js";
import { exportCookies } from "./components/cookies.js";
import { exportMap } from "./components/delivery-map.js";
import { animateScrollToAnchor } from "./components/animateScrollToAnchor.js";
import "./components/faq.js";

exportHeader();
exportCookies();
exportMap();

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
}
