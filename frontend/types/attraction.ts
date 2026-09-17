export interface AttractionPhoto { id: number; position: number }
export interface AttractionSubcategory { id: number; name: string; category_id: number }
export interface AttractionCategory { id: number; name: string; subcategories: AttractionSubcategory[] }
export interface AttractionClassification { subcategory_id: number; position: number; subcategory: AttractionSubcategory }
export interface Attraction {
  subcategories: AttractionClassification[];
  schedule_mode: 'unspecified' | 'scheduled' | 'all_day'; opening_time: string; closing_time: string; opening_days: number[];
  season_mode: 'unspecified' | 'all_year' | 'months'; season_start_month: number | null; season_end_month: number | null;
  id: number; name: string; description: string; category: string; department: string; city: string; address: string;
  latitude: number | null; longitude: number | null; opening_hours: string; admission_cents: number; recommendations: string; phone: string;
  photos: AttractionPhoto[]; updated_at: string;
  manager_id?: number; manager_name?: string; manager_email?: string; manager_status?: string; status?: string; published?: boolean; version?: number;
}
export interface AttractionOptions { departments: string[]; categories: AttractionCategory[] }
export interface AttractionPage { attractions: Attraction[]; pagination: { total: number } }
