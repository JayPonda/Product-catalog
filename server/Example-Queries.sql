
# Given a table of Orders and a table of Customers, 
# write a query to retrieve the top 5 customers with the highest total order value.

# You have a table with Orders (OrderId, CustomerId, OrderDate, TotalAmount). 

# solution 1: if you need 5 heightest paying customer (same order number will considered aligible)
With order_per_customer as (
    SELECT o.customer_id, sum(o.total_bill ) AS total_cost  FROM orders o GROUP BY o.customer_id 
), top_orders as (
    SELECT opc.total_cost AS top_cost FROM order_per_customer opc GROUP BY opc.total_cost ORDER BY opc.total_cost DESC limit 5
) SELECT * FROM users c
JOIN order_per_customer opc 
ON c.id = opc.customer_id
JOIN top_orders 
ON opc.total_cost = top_orders.top_cost;


# solution 2: if you need 5 customers no matter the order cost
SELECT * FROM customers c
JOIN (
    SELECT o.customer_id, sum(o.total_amount) AS total_cost  FROM orders o GROUP BY o.customer_id ORDER BY sum(o.total_amount) DESC limit 5
) top_order_customers 
ON c.id = top_order_customers.customer_id
ORDER BY top_order_customers.total_cost DESC;


# Write a query to get the total amount spent by each customer in the current month, 
# ordered by the highest spenders first.

SELECT o.customer_id, sum(o.total_amount) AS total_cost  
FROM orders o 
WHERE o.order_date >= DATE_TRUNC('month', CURRENT_DATE) 
GROUP BY o.customer_id 
ORDER BY sum(o.total_amount) DESC;


# Write a query to get the latest orders from each customer. 
# The query should return all the
# columns (OrderId, CustomerId, OrderDate, TotalAmount) from the Orders table for
# this list of latest orders.

-- this query may looks correct but this query is not a right as this may produce the wrong result.
-- here, order placing is human and human have limitation fo the behavior so 
-- i can tell you this query works but for true solution you need to sitroduce some other column which may helping to clear this.
-- we use uuid, so it can't be shorted just like a sequance does, so we need to use sequance column as well. here. 
-- if database can distinquash the two records this qury works seemless. 
WITH latest_order_of_customer as (
    SELECT o.customer_id, max(o.created_at) as latest_order_time FROM orders o GROUP o.customer_id -- can generate more then one record per customer (example seed data by script)
), customer_letest_record as (
    SELECT o.id, looc.customer_id, looc.latest_order_time, o.total_bill FROM orders o JOIN latest_order_of_customer looc ON looc.customer_id = o.customer_id AND looc.latest_order_time = o.created_at -- can have multiple one per customer
) SELECT u.first_name, clr.* FROM customer_letest_record clr JOIN users u ON clr.customer_id = u.id

# Optimization Query: Can you identify potential issues and suggest improvements
# for the query below?
# SELECT *
# FROM Orders
#WHERE CustomerId IN (SELECT CustomerId FROM Customers WHERE Country = 'USA');

-- solutions
-- 1. don't use star if possible.
-- 2. here we use "In" this can be avoided by joins, dbs are optimised for that.
-- 3. this query is written to get orders, nice but this can leads to hundrads of records, so use limit/offset as well.
-- 4. if possible always carry the limit of data, like for just this months or this date (limit it upto the actual need)
-- 5. i hope there is an index for country column for best performance.