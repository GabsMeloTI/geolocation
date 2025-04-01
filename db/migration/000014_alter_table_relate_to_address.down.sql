ALTER TABLE public.neighborhoods DROP COLUMN IF EXISTS lat;
ALTER TABLE public.neighborhoods DROP COLUMN IF EXISTS lon;

ALTER TABLE public.states DROP COLUMN IF EXISTS lat;
ALTER TABLE public.states DROP COLUMN IF EXISTS lon;

ALTER TABLE public.cities DROP COLUMN IF EXISTS lat;
ALTER TABLE public.cities DROP COLUMN IF EXISTS lon;
