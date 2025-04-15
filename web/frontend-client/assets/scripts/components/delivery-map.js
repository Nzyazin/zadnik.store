export function exportMap () {
  if (document.getElementById('delivery-map')) {
    setTimeout(initMap, 0)
  }
}

function initMap () {

  const centerMap = [58.684647, 50.025364]
  const zoom = window.innerWidth > 950 ? 5 : 3
  const contactMapControls = window.innerWidth > 950 ? [] : ['zoomControl']
  const contactMapBehaviors = window.innerWidth > 950 ? ['drag', 'dblClickZoom', 'multiTouch'] : ['dblClickZoom', 'multiTouch']

  let scriptTag

  createAndLoadScriptMap()

  function createAndLoadScriptMap () {
    scriptTag = document.createElement('script')
    scriptTag.async = true
    scriptTag.src = 'https://api-maps.yandex.ru/2.1/?lang=ru_RU'
    scriptTag.addEventListener('load', createMap)
    document.body.appendChild(scriptTag)

  }

  function createMap () {
    scriptTag.removeEventListener('load', createMap)
    ymaps.ready(initMap)
  }

  function initMap () {
    let map = new ymaps.Map('delivery-map', {
      center: centerMap,
      zoom: zoom,
      controls: contactMapControls,
      behaviors: contactMapBehaviors
    })

    const vahrushi = new ymaps.Placemark([58.684647, 50.025364], {
      iconCaption: 'п. Вахруши, ул. Рабочая 30'
    })
    map.geoObjects.add(vahrushi)
  }
}
