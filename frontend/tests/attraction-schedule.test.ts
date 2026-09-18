import assert from 'node:assert/strict'
import test from 'node:test'
import { scheduleLabel, seasonLabel } from '../utils/attractions.ts'

test('visit hours preserve local times and explain overnight closing', () => {
  assert.equal(scheduleLabel({ schedule_mode: 'scheduled', opening_time: '22:00', closing_time: '02:15', opening_days: [7, 1], opening_hours: '' }), 'Lunes, Domingo · 22:00 a 02:15 (cierre al día siguiente)')
  assert.equal(scheduleLabel({ schedule_mode: 'all_day', opening_time: '', closing_time: '', opening_days: [1,2,3,4,5,6,7], opening_hours: '' }), 'Todos los días · 24 horas')
  assert.equal(scheduleLabel({ schedule_mode: 'unspecified', opening_time: '', closing_time: '', opening_days: [], opening_hours: 'Horario anterior conservado' }), 'Horario anterior conservado')
})
test('seasons support year boundaries and a single month', () => {
  assert.equal(seasonLabel({ season_mode: 'months', season_start_month: 11, season_end_month: 2 }), 'Noviembre a Febrero (hasta el año siguiente)')
  assert.equal(seasonLabel({ season_mode: 'months', season_start_month: 3, season_end_month: 3 }), 'Marzo')
  assert.equal(seasonLabel({ season_mode: 'all_year', season_start_month: null, season_end_month: null }), 'Todo el año')
})
