Level 2 for completing a technical school assignment.

Все задания находятся в разных ветках, которые помечены своими номерами, соответствующие номерам в курсе.
Это ветка с заданнием L 2.18.

Порядок:
1) go run main.go
2) Invoke-RestMethod -Uri http://localhost:8080/create_event `
>>   -Method POST `
>>   -Body (@{ user_id = 1; date = '2025-08-11'; event = 'Test event' } | ConvertTo-Json) `
>>   -ContentType "application/json"
3)  Invoke-RestMethod -Uri "http://localhost:8080/events_for_day?user_id=1&date=2025-08-11" -Method GET
4) Invoke-RestMethod -Uri http://localhost:8080/update_event `                                        
>>   -Method POST `
>>   -Body @{ id = 1; user_id = 1; date = '2025-08-12'; event = 'Updated event' }
5) Invoke-RestMethod -Uri "http://localhost:8080/events_for_day?user_id=1&date=2025-08-12" -Method GET
6) Invoke-RestMethod -Uri http://localhost:8080/delete_event `                                        
>>   -Method POST `
>>   -Body @{ id = 1 }
7) Invoke-RestMethod -Uri "http://localhost:8080/events_for_day?user_id=1&date=2025-08-12" -Method GET
