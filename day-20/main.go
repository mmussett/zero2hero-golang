package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type Task struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

func dataPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".todo.json")
}

func load() ([]Task, error) {
	data, err := os.ReadFile(dataPath())
	if os.IsNotExist(err) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []Task
	return tasks, json.Unmarshal(data, &tasks)
}

func save(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataPath(), data, 0o644)
}

func nextID(tasks []Task) int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}

func main() {
	var showAll bool

	root := &cobra.Command{Use: "todo", Short: "A simple todo list manager"}

	root.AddCommand(&cobra.Command{
		Use:   "add [task text]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := load()
			if err != nil {
				return err
			}
			t := Task{ID: nextID(tasks), Text: args[0], CreatedAt: time.Now()}
			tasks = append(tasks, t)
			if err := save(tasks); err != nil {
				return err
			}
			fmt.Printf("Added task #%d: %q\n", t.ID, t.Text)
			return nil
		},
	})

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := load()
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Println("No tasks.")
				return nil
			}
			for _, t := range tasks {
				if !showAll && t.Done {
					continue
				}
				status := "[ ]"
				if t.Done {
					status = "[x]"
				}
				fmt.Printf("%s #%-3d %s\n", status, t.ID, t.Text)
			}
			return nil
		},
	}
	listCmd.Flags().BoolVar(&showAll, "all", false, "Show completed tasks too")
	root.AddCommand(listCmd)

	root.AddCommand(&cobra.Command{
		Use:   "done [id]",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}
			tasks, err := load()
			if err != nil {
				return err
			}
			for i, t := range tasks {
				if t.ID == id {
					tasks[i].Done = true
					if err := save(tasks); err != nil {
						return err
					}
					fmt.Printf("Marked #%d as done: %q\n", id, t.Text)
					return nil
				}
			}
			return fmt.Errorf("task #%d not found", id)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}
			tasks, err := load()
			if err != nil {
				return err
			}
			for i, t := range tasks {
				if t.ID == id {
					tasks = append(tasks[:i], tasks[i+1:]...)
					if err := save(tasks); err != nil {
						return err
					}
					fmt.Printf("Deleted #%d: %q\n", id, t.Text)
					return nil
				}
			}
			return fmt.Errorf("task #%d not found", id)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "export",
		Short: "Export tasks as CSV to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := load()
			if err != nil {
				return err
			}
			w := csv.NewWriter(os.Stdout)
			w.Write([]string{"id", "text", "done", "created_at"})
			for _, t := range tasks {
				w.Write([]string{
					strconv.Itoa(t.ID),
					t.Text,
					strconv.FormatBool(t.Done),
					t.CreatedAt.Format(time.RFC3339),
				})
			}
			w.Flush()
			return w.Error()
		},
	})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
