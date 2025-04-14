export function exportDeliveryMap () {
  if (document.getElementById('delivery-map')) {
    setTimeout(initDeliveryMap, 0)
  }
}

function initDeliveryMap () {

  const centerMap = [56.996127, 46.106405]
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

    const kirov = new ymaps.Placemark([58.603591, 49.668014], {
      iconCaption: 'Орловская, 4Г'
    })
    map.geoObjects.add(kirov)

    const moscow = new ymaps.Placemark([55.709312, 37.798017], {
      iconCaption: 'Москва, Сормовский проезд, 7А'
    })
    map.geoObjects.add(moscow)
  }
}
