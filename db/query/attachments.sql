-- name: CreateAttachments :one
INSERT INTO public.attachments
(id, user_id, description, url, name_file, size_file, status, created_at)
VALUES(nextval('attachments_id_seq'::regclass), $1, $2, $3, $4, $5, true, now())
    RETURNING *;

-- name: GetAttachmentById :one
SELECT *
FROM public.attachments
WHERE id=$1;

-- name: UpdateAttachmentLogicDelete :exec
UPDATE public.attachments
SET status=false, updated_at=now()
WHERE id = $1;


