from datetime import date
import unittest
from check_npm import check


class AuditGateTest(unittest.TestCase):
    def setUp(self):
        self.audit = {'auditReportVersion': 2, 'vulnerabilities': {
            'tool': {'nodes': ['node_modules/tool'], 'via': [
                {'url': 'https://github.com/advisories/GHSA-test', 'severity': 'high'}]},
            'parent': {'nodes': [], 'via': ['tool']}}}
        self.lock = {'packages': {'node_modules/tool': {'version': '1.0.0'}}}
        self.exception = [{'module': 'tool', 'advisory': 'GHSA-test', 'versions': ['1.0.0'],
                           'expires': '2026-11-01', 'reason': 'build-only', 'owner': 'maintainer'}]

    def test_unknown_advisory_blocks(self):
        errors, _ = check(self.audit, self.lock, set(), [], date(2026, 10, 23))
        self.assertEqual(len(errors), 1)

    def test_only_reviewed_build_version_is_accepted(self):
        errors, accepted = check(self.audit, self.lock, set(), self.exception, date(2026, 10, 23))
        self.assertFalse(errors)
        self.assertEqual(len(accepted), 1)
        self.lock['packages']['node_modules/tool']['version'] = '1.0.1'
        errors, _ = check(self.audit, self.lock, set(), self.exception, date(2026, 10, 23))
        self.assertTrue(errors)

    def test_shipped_or_expired_exception_blocks(self):
        for runtime, today in [({'tool'}, date(2026, 10, 23)), (set(), date(2026, 11, 2))]:
            errors, _ = check(self.audit, self.lock, runtime, self.exception, today)
            self.assertTrue(errors)

    def test_scanner_error_cannot_pass(self):
        with self.assertRaises(ValueError):
            check({'error': {}}, self.lock, set(), [], date(2026, 10, 23))


if __name__ == '__main__':
    unittest.main()
