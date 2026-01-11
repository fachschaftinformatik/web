-- +goose Up
-- +goose StatementBegin
DELETE FROM modules 
WHERE name LIKE '%Bachelorarbeit%' 
   OR name LIKE '%Masterarbeit%' 
   OR name LIKE '%Praxisphase%' 
   OR name LIKE '%Projekt%' 
   OR name LIKE 'Kolloquium%';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Re-insertion would be complex due to programid dependencies, omitting for now
-- as migrations are usually focused on the forward path in this dev environment.
-- +goose StatementEnd
