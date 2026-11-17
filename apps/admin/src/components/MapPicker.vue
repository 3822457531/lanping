<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import http from '@/api/http'
import { gcj02ToWgs84, wgs84ToGcj02 } from '@/utils/geo'

const props = withDefaults(
  defineProps<{
    modelValue?: boolean
    lat?: number
    lng?: number
    rangeM?: number
    addressHint?: string
  }>(),
  {
    modelValue: false,
    lat: 0,
    lng: 0,
    rangeM: 200,
    addressHint: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [boolean]
  confirm: [payload: { lat: number; lng: number; address?: string }]
}>()

const mapEl = ref<HTMLElement | null>(null)
const searchQ = ref('')
const searching = ref(false)
const tileOk = ref(true)
const picked = ref<{ lat: number; lng: number } | null>(null) // 始终存 WGS84
const searchHits = ref<{ display: string; lat: number; lng: number }[]>([])

let map: L.Map | null = null
let marker: L.Marker | null = null
let circle: L.Circle | null = null
let tileLayer: L.TileLayer | null = null

// 高德瓦片（国内可访问）为 GCJ-02；落库/打卡用 WGS84
const defaultWgs: [number, number] = [22.5431, 114.0579]

function close() {
  emit('update:modelValue', false)
}

function setPointWgs(lat: number, lng: number, fly = true) {
  picked.value = { lat, lng }
  if (!map) return
  const g = wgs84ToGcj02(lat, lng)
  if (marker) marker.setLatLng([g.lat, g.lng])
  else {
    const icon = L.divIcon({
      className: 'lp-map-pin',
      html: '<span></span>',
      iconSize: [22, 22],
      iconAnchor: [11, 11],
    })
    marker = L.marker([g.lat, g.lng], { draggable: true, icon }).addTo(map)
    marker.on('dragend', () => {
      const p = marker!.getLatLng()
      const w = gcj02ToWgs84(p.lat, p.lng)
      setPointWgs(w.lat, w.lng, false)
    })
  }
  const radius = Math.max(50, Number(props.rangeM) || 200)
  if (circle) {
    circle.setLatLng([g.lat, g.lng])
    circle.setRadius(radius)
  } else {
    circle = L.circle([g.lat, g.lng], {
      radius,
      color: '#e24b6e',
      fillColor: '#e24b6e',
      fillOpacity: 0.12,
      weight: 2,
    }).addTo(map)
  }
  if (fly) map.flyTo([g.lat, g.lng], Math.max(map.getZoom(), 16), { duration: 0.45 })
}

function initMap() {
  if (!mapEl.value || map) return
  tileOk.value = true
  const has = !!(props.lat && props.lng)
  const centerWgs = has ? { lat: props.lat, lng: props.lng } : { lat: defaultWgs[0], lng: defaultWgs[1] }
  const centerG = wgs84ToGcj02(centerWgs.lat, centerWgs.lng)

  map = L.map(mapEl.value, { zoomControl: true }).setView([centerG.lat, centerG.lng], has ? 16 : 12)

  // 高德矢量路网（国内可达）；OSM 在国内常超时导致灰屏
  tileLayer = L.tileLayer(
    'https://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}',
    {
      subdomains: '1234',
      maxZoom: 18,
      attribution: '© 高德地图',
    },
  )
  let tileErrorCount = 0
  tileLayer.on('tileerror', () => {
    tileErrorCount += 1
    if (tileErrorCount >= 4) tileOk.value = false
  })
  tileLayer.on('load', () => {
    tileOk.value = true
  })
  tileLayer.addTo(map)

  map.on('click', (e: L.LeafletMouseEvent) => {
    const w = gcj02ToWgs84(e.latlng.lat, e.latlng.lng)
    setPointWgs(w.lat, w.lng, false)
  })

  if (has) setPointWgs(props.lat, props.lng, false)
  setTimeout(() => map?.invalidateSize(), 120)
  setTimeout(() => map?.invalidateSize(), 360)
}

async function search() {
  const q = searchQ.value.trim() || props.addressHint.trim()
  if (!q) return ElMessage.warning('请输入地址关键词')
  searching.value = true
  searchHits.value = []
  try {
    const data: any = await http.get('/api/geo/search', { params: { q } })
    const list = data.list || []
    searchHits.value = list.map((x: any) => ({
      display: x.name || x.address || '',
      lat: Number(x.lat),
      lng: Number(x.lng),
    }))
    if (!searchHits.value.length) {
      ElMessage.warning(data.message || '未找到地址，可直接在地图上点击选点')
      return
    }
    const hit = searchHits.value[0]
    setPointWgs(hit.lat, hit.lng, true)
  } catch (e: any) {
    ElMessage.error(e.message || '地址搜索失败（可点击地图选点）')
  } finally {
    searching.value = false
  }
}

function pickHit(hit: { lat: number; lng: number; display: string }) {
  setPointWgs(hit.lat, hit.lng, true)
  searchQ.value = hit.display
}

function useBrowserLocation() {
  if (!navigator.geolocation) return ElMessage.warning('当前浏览器不支持定位')
  navigator.geolocation.getCurrentPosition(
    (pos) => setPointWgs(pos.coords.latitude, pos.coords.longitude, true),
    () => ElMessage.error('定位失败，请允许位置权限'),
    { enableHighAccuracy: true, timeout: 12000 },
  )
}

function confirm() {
  if (!picked.value) return
  emit('confirm', {
    lat: Number(picked.value.lat.toFixed(7)),
    lng: Number(picked.value.lng.toFixed(7)),
    address: searchQ.value.trim() || undefined,
  })
  close()
}

function destroyMap() {
  if (map) {
    map.remove()
    map = null
  }
  marker = null
  circle = null
  tileLayer = null
}

watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      searchQ.value = props.addressHint || ''
      picked.value = props.lat && props.lng ? { lat: props.lat, lng: props.lng } : null
      await nextTick()
      initMap()
    } else {
      destroyMap()
      searchHits.value = []
    }
  },
)

watch(
  () => props.rangeM,
  (r) => {
    if (circle) circle.setRadius(Math.max(50, Number(r) || 200))
  },
)

onBeforeUnmount(destroyMap)
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    title="地图选点 · 打卡中心"
    width="760px"
    top="6vh"
    destroy-on-close
    append-to-body
    @update:model-value="emit('update:modelValue', $event)"
    @opened="() => map?.invalidateSize()"
  >
    <div class="bar">
      <el-input
        v-model="searchQ"
        clearable
        placeholder="输入地址搜索，或直接在地图上点击"
        @keyup.enter="search"
      />
      <el-button type="primary" :loading="searching" @click="search">搜索</el-button>
      <el-button @click="useBrowserLocation">用我当前位置</el-button>
    </div>
    <div v-if="searchHits.length" class="hits">
      <button v-for="(h, i) in searchHits" :key="i" type="button" class="hit" @click="pickHit(h)">
        {{ h.display }}
      </button>
    </div>
    <div v-if="!tileOk" class="tile-warn">
      底图加载异常。请检查网络；仍可点击空白区域选点（粉圈为打卡范围）。
    </div>
    <div ref="mapEl" class="map" />
    <div class="meta">
      <div>
        已选（WGS84 打卡坐标）：
        <template v-if="picked">
          {{ picked.lat.toFixed(6) }}, {{ picked.lng.toFixed(6) }}
        </template>
        <span v-else class="muted">点击地图选点（粉圈约 {{ rangeM }} 米）</span>
      </div>

    </div>
    <template #footer>
      <el-button @click="close">取消</el-button>
      <el-button type="primary" :disabled="!picked" @click="confirm">确认选点</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.bar {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}
.hits {
  display: grid;
  gap: 4px;
  margin-bottom: 8px;
  max-height: 110px;
  overflow: auto;
}
.hit {
  text-align: left;
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 12px;
  color: #344054;
  cursor: pointer;
}
.hit:hover { border-color: var(--el-color-primary-light-5); }
.tile-warn {
  margin-bottom: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 12px;
}
.map {
  height: 420px;
  width: 100%;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--el-border-color);
  z-index: 0;
  background: #e8eef5;
}
.meta {
  margin-top: 10px;
  font-size: 13px;
  display: grid;
  gap: 4px;
}
.muted { color: var(--el-text-color-secondary); font-size: 12px; }
</style>

<style>
.lp-map-pin {
  background: transparent !important;
  border: 0 !important;
}
.lp-map-pin span {
  display: block;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #e24b6e;
  border: 3px solid #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.28);
}
</style>
