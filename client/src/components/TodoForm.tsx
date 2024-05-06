import { Button, Flex, Input, Spinner } from "@chakra-ui/react";
import React, { useState } from "react";
import { IoMdAdd } from "react-icons/io";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { BASE_URL } from "../App";

const TodoForm = () => {
	const [newTodo, setNewTodo] = useState("");
	const queryClient = useQueryClient();

	const {mutate:createTodo,isPending:isCreating} = useMutation({
		mutationKey: ["createTodo"],
		mutationFn: async (e: React.FormEvent) => {
			e.preventDefault()
			if(!newTodo.trim()){
				return alert("Todo can't be empty")
			}
			try {
				const response = await fetch(BASE_URL+ `/todos`,{
					method:"POST",
					headers:{
						"Content-Type":"application/json",
					},
					body:JSON.stringify({body:newTodo}),
				})
				const data = await response.json()
				if (!response.ok) {
					throw new Error(data.error || "Something went wrong!");
				}
				setNewTodo("")
				return data
			} catch (error) {
				console.log(error)
			}
		},
		onSuccess: () => {
			queryClient.invalidateQueries({queryKey:["todos"]})
		},
		onError: (error) => {
			alert(error.message)
		},
	})
	return (
		<form onSubmit={createTodo}>
			<Flex gap={2}>
				<Input
					type='text'
					value={newTodo}
					onChange={(e) => setNewTodo(e.target.value)}
					ref={(input) => input && input.focus()}
				/>
				<Button
					mx={2}
					type='submit'
					_active={{
						transform: "scale(.97)",
					}}
				>
					{isCreating ? <Spinner size={"md"} /> : <IoMdAdd size={30} />}
				</Button>
			</Flex>
		</form>
	);
};
export default TodoForm;